// uploader-submitter.mjs

const arc = "application/x-freearc",
  bz = "application/x-bzip",
  bz2 = "application/x-bzip2",
  gzip = "application/gzip",
  rar = "application/vnd.rar",
  tar = "application/x-tar",
  zip = "application/zip",
  zip7 = "application/x-7z-compressed";

const dos = "application/x-msdos-program";

const gif = "image/gif",
  jpeg = "image/jpeg",
  png = "image/png";

export function apps() {
  const allowedTypes = [dos];
  return allowedTypes;
}

export function archives() {
  const allowedTypes = [arc, bz, bz2, gzip, rar, tar, zip, zip7];
  return allowedTypes;
}

export function binaries() {
  const allowedTypes = ["application/octet-stream", "application/x-binary"];
  return allowedTypes;
}

export function images() {
  const allowedTypes = [gif, jpeg, png];
  return allowedTypes;
}

export function texts() {
  const allowedTypes = ["text/plain"];
  return allowedTypes;
}

export function checkImage(mime) {
  const allowedTypes = images().concat(archives());
  return allowedTypes.includes(mime);
}

export function checkIntro(mime) {
  const allowedTypes = apps().concat(archives(), binaries());
  return allowedTypes.includes(mime);
}

export function checkMagazine(mime) {
  const allowedTypes = texts().concat(archives(), apps(), binaries());
  return allowedTypes.includes(mime);
}

export function checkText(mime) {
  const allowedTypes = texts().concat(archives());
  return allowedTypes.includes(mime);
}
