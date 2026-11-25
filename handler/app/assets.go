//nolint:exhaustruct_v5,gochecknoglobals
package app

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"os"
	"path/filepath"
	"runtime/metrics"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/dustin/go-humanize"
)

// AssetCache holds the file content stats for an asset directory.
type AssetCache struct {
	sync.RWMutex

	c    int
	b    int64
	exts []helper.Extension
}

var (
	downloadCache  = &AssetCache{c: -1, b: -1}
	previewCache   = &AssetCache{c: -1, b: -1}
	thumbnailCache = &AssetCache{c: -1, b: -1}
	extraCache     = &AssetCache{c: -1, b: -1}
	orphanCache    = &AssetCache{c: -1, b: -1}
)

type assetSpec struct {
	path   string
	desc   string
	prefix string
	cache  *AssetCache
}

func assetSpecs(conf config.Config) []assetSpec {
	return []assetSpec{
		{path: string(conf.AbsDownload), desc: "downloads", prefix: "Downloads", cache: downloadCache},
		{path: conf.AbsPreview.String(), desc: "previews", prefix: "Previews", cache: previewCache},
		{path: conf.AbsThumbnail.String(), desc: "thumbnails", prefix: "Thumbnails", cache: thumbnailCache},
		{path: conf.AbsExtra.String(), desc: "extra", prefix: "Extras", cache: extraCache},
		{path: conf.AbsOrphaned.String(), desc: "orphaned", prefix: "Orphaned", cache: orphanCache},
	}
}

// AssetsCached reports whether all five asset directory caches have already been populated.
func AssetsCached() bool {
	for _, c := range []*AssetCache{downloadCache, previewCache, thumbnailCache, extraCache, orphanCache} {
		c.RLock()
		warm := c.c != -1
		c.RUnlock()
		if !warm {
			return false
		}
	}
	return true
}

// HostStats collects lightweight host disk, memory, and Go runtime heap metrics concurrently.
func HostStats(conf config.Config) map[string]any {
	data := make(map[string]any)
	var wg sync.WaitGroup
	var mu sync.Mutex

	disks := [...]struct{ prefix, path string }{
		{"wd", conf.AbsDownload.String()},
		{"tmp", os.TempDir()},
		{"os", "/"},
	}
	for _, d := range disks {
		wg.Go(func() {
			total, free, _, pct, _ := helper.DiskStat(d.path)
			mu.Lock()
			data[d.prefix+"Total"] = helper.ByteCount(int64(total))
			data[d.prefix+"Free"] = helper.ByteCount(int64(free))
			data[d.prefix+"Percent"] = pct
			mu.Unlock()
		})
	}

	const unit = 1024
	wg.Go(func() {
		samples := []metrics.Sample{
			{Name: "/memory/classes/heap/objects:bytes"}, // Live heap objects
			{Name: "/memory/classes/total:bytes"},        // Total memory managed by runtime
		}
		metrics.Read(samples)
		m0 := samples[0].Value.Uint64() / (unit * unit)
		m1 := samples[1].Value.Uint64() / (unit * unit)
		mu.Lock()
		data["metric0"] = m0
		data["metric1"] = m1
		mu.Unlock()
	})

	wg.Go(func() { memInfo(&mu, data) })

	wg.Wait()
	return data
}

// AssetDirsData collects file counts, sizes, and extension breakdowns for all five asset directories
// concurrently. When refresh is true, cached directory stats are invalidated and rescanned.
func AssetDirsData(conf config.Config, refresh bool) map[string]any {
	data := make(map[string]any)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, spec := range assetSpecs(conf) {
		wg.Go(func() {
			part := processAsset(spec, refresh)
			mu.Lock()
			maps.Copy(data, part)
			mu.Unlock()
		})
	}

	wg.Wait()
	return data
}

// AssetsData collects both host disk/memory telemetry and all asset directory stats concurrently.
func AssetsData(conf config.Config, refresh bool) map[string]any {
	data := make(map[string]any)
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Go(func() {
		host := HostStats(conf)
		mu.Lock()
		maps.Copy(data, host)
		mu.Unlock()
	})

	wg.Go(func() {
		dirs := AssetDirsData(conf, refresh)
		mu.Lock()
		maps.Copy(data, dirs)
		mu.Unlock()
	})

	wg.Wait()
	return data
}

// processAsset validates the directory, reads or refreshes its AssetCache in a single FS pass,
// and returns the template key/value map without holding any shared map lock during I/O.
func processAsset(spec assetSpec, refresh bool) map[string]any {
	countKey := "count" + spec.prefix
	usageKey := "usage" + spec.prefix
	extsKey := "exts" + spec.prefix
	checkKey := "check" + spec.prefix

	out := map[string]any{
		countKey: 0,
		usageKey: 0,
		extsKey:  []helper.Extension{},
	}

	name := dir.Directory(spec.path)
	check := config.CheckDir(name, spec.desc)
	out[checkKey] = check
	if check != nil {
		return out
	}

	if !refresh {
		spec.cache.RLock()
		if spec.cache.c != -1 {
			c, b, exts := spec.cache.c, spec.cache.b, spec.cache.exts
			spec.cache.RUnlock()

			out[countKey] = humanize.Comma(int64(c))
			out[usageKey] = helper.ByteCountFloat(b)
			out[extsKey] = exts
			return out
		}
		spec.cache.RUnlock()
	}

	spec.cache.Lock()
	defer spec.cache.Unlock()

	if refresh || spec.cache.c == -1 {
		c, b, exts, _ := scanAssetDir(spec.path)
		spec.cache.c = c
		spec.cache.b = b
		spec.cache.exts = exts
	}

	out[countKey] = humanize.Comma(int64(spec.cache.c))
	out[usageKey] = helper.ByteCountFloat(spec.cache.b)
	out[extsKey] = spec.cache.exts
	return out
}

type dirScan struct {
	count     int
	total     int64
	extCounts map[string]int64
	queue     []string
}

func (d *dirScan) recordEntry(cur string, entry os.DirEntry) {
	name := entry.Name()
	if entry.IsDir() {
		d.queue = append(d.queue, filepath.Join(cur, name))
		return
	}
	if !entry.Type().IsRegular() || name == helper.DSStore {
		return
	}
	d.count++
	if info, err := entry.Info(); err == nil {
		d.total += info.Size()
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		ext = "uuid"
	}
	d.extCounts[ext]++
}

func (d *dirScan) scanDir(cur string, isRoot bool) error {
	const batchSize = 1024
	f, err := os.Open(cur)
	if err != nil {
		if isRoot {
			return fmt.Errorf("scan asset dir open %s: %w", cur, err)
		}
		return nil
	}
	defer func() { _ = f.Close() }()

	for {
		entries, readErr := f.ReadDir(batchSize)
		for _, entry := range entries {
			d.recordEntry(cur, entry)
		}
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("scan asset dir read %s: %w", cur, readErr)
		}
	}
}

// scanAssetDir computes file count, total bytes, and extension breakdown in a single
// streaming pass without sorting directory entries in memory.
// The returned values are the count, total, and sorted extensions.
func scanAssetDir(root string) (int, int64, []helper.Extension, error) {
	d := dirScan{
		extCounts: make(map[string]int64),
		queue:     []string{root},
	}
	for len(d.queue) > 0 {
		cur := d.queue[0]
		d.queue = d.queue[1:]
		if err := d.scanDir(cur, cur == root); err != nil {
			if d.count == 0 && len(d.extCounts) == 0 {
				return 0, 0, nil, err
			}
			return d.count, d.total, sortExtensions(d.extCounts), err
		}
	}
	return d.count, d.total, sortExtensions(d.extCounts), nil
}

func sortExtensions(extCounts map[string]int64) []helper.Extension {
	exts := make([]helper.Extension, 0, len(extCounts))
	for k, v := range extCounts {
		exts = append(exts, helper.Extension{Name: k, Count: v})
	}
	slices.SortFunc(exts, func(a, b helper.Extension) int {
		if a.Count != b.Count {
			return cmp.Compare(b.Count, a.Count)
		}
		return cmp.Compare(a.Name, b.Name)
	})
	return exts
}

func memInfo(mu *sync.Mutex, data map[string]any) {
	const (
		na   = "n/a"
		unit = 1024
		calc = float64(100.0)
		req  = 2
	)
	vals := map[string]any{
		"memTotal": na, "memFree": na, "memPercent": "0%",
		"swapTotal": na, "swapFree": na, "swapPercent": "0%",
	}
	defer func() {
		mu.Lock()
		maps.Copy(data, vals)
		mu.Unlock()
	}()
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()
	mb := func(s string) string {
		kb, _ := strconv.Atoi(s)
		return fmt.Sprintf("%d MB", kb/unit)
	}
	perc := func(f, t string) string {
		free, _ := strconv.Atoi(f)
		total, _ := strconv.Atoi(t)
		if total == 0 {
			return "0%"
		}
		return fmt.Sprintf("%d%%", int64(math.Round((float64(free)/float64(total))*calc)))
	}
	raw := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := strings.Fields(scanner.Text())
		if len(s) >= req {
			raw[s[0]] = s[1]
		}
	}
	if err := scanner.Err(); err != nil {
		return
	}
	// Prefer MemAvailable on Linux (includes reclaimable page cache) over raw MemFree.
	memAvail := raw["MemAvailable:"]
	if memAvail == "" {
		memAvail = raw["MemFree:"]
	}
	if v, ok := raw["MemTotal:"]; ok {
		vals["memTotal"] = mb(v)
	}
	if memAvail != "" {
		vals["memFree"] = mb(memAvail)
	}
	if v, ok := raw["SwapTotal:"]; ok {
		vals["swapTotal"] = mb(v)
	}
	if v, ok := raw["SwapFree:"]; ok {
		vals["swapFree"] = mb(v)
	}
	vals["memPercent"] = perc(memAvail, raw["MemTotal:"])
	vals["swapPercent"] = perc(raw["SwapFree:"], raw["SwapTotal:"])
}
