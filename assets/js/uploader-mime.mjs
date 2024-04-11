// uploader-submitter.mjs

import { getElmById } from "./helper.mjs";
export default intro;

const arc = "application/x-freearc",
  binary = "application/octet-stream",
  bz = "application/x-bzip",
  bz2 = "application/x-bzip2",
  gzip = "application/gzip",
  rar = "application/vnd.rar",
  tar = "application/x-tar",
  zip = "application/zip",
  zip7 = "application/x-7z-compressed";

export function archive(mime) {
  const allowedTypes = [arc, bz, bz2, gzip, rar, tar, zip, zip7];
  if (allowedTypes.includes(mime) == false) {
    return false;
  }
  return true;
}

export function unknown(mime) {
  const allowedTypes = [binary];
  return allowedTypes.includes(mime);
}

export function intro(mime) {
  ok = false;
  ok = unknown(mime);
  if (ok) {
    return true;
  }
  ok = archive(mime);
  if (ok) {
    return true;
  }
  return false;
}
