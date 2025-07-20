/**
 * @module uploader-mime
 * This module provides functions for handling file uploads mime types.
 */
const arc = "application/x-freearc",
  arj = "application/x-arj",
  bz = "application/x-bzip",
  bz2 = "application/x-bzip2",
  gzip = "application/gzip",
  rar = "application/vnd.rar",
  tar = "application/x-tar",
  zip = "application/zip",
  zipx = "application/x-zip-compressed",
  zip7 = "application/x-7z-compressed";

const dos = "application/x-msdos-program";

const bmp = "image/bmp",
  gif = "image/gif",
  jpeg = "image/jpeg",
  pcx = "image/vnd.zbrush.pcx",
  png = "image/png",
  tiff = "image/tiff",
  webp = "image/webp";

const csh = "application/x-csh",
  ext = "application/x-chrome-extension",
  perl = "text/x-script.perl",
  php = "application/x-httpd-php",
  py = "text/x-script.phyton",
  rexx = "text/x-script.rexx",
  sh = "application/x-sh",
  ssh = "application/x-shellscript",
  tcl = "text/x-script.tcl",
  xsh = "text/x-shellscript",
  zsh = "text/x-script.zsh";

export function reject() {
  const types = [csh, ext, perl, php, py, rexx, sh, ssh, tcl, xsh, zsh];
  return types;
}

export function apps() {
  const allowedTypes = [dos];
  return allowedTypes;
}

export function archives() {
  const allowedTypes = [arc, arj, bz, bz2, gzip, rar, tar, zip, zipx, zip7];
  return allowedTypes;
}

export function binaries() {
  const allowedTypes = [
    "application/octet-stream",
    "application/x-binary",
    "application/x-ms-dos-executable",
  ];
  return allowedTypes;
}

export function images() {
  const allowedTypes = [bmp, gif, jpeg, pcx, png, tiff, webp];
  return allowedTypes;
}

export function texts() {
  const allowedTypes = ["text/plain", "text/x-nfo"];
  return allowedTypes;
}

/**
 * Checks if the given MIME type is in the list of rejected types.
 *
 * @param {string} mime - The MIME type to check.
 * @returns {boolean} - Returns true if the MIME type is in the list of rejected types, otherwise false.
 */
export function checkAdvanced(mime) {
  const rejectTypes = reject();
  return rejectTypes.includes(mime);
}

/**
 * Checks if the given MIME type is allowed for images.
 * @param {string} mime - The MIME type to check.
 * @returns {boolean} - Returns true if the MIME type is allowed for images, otherwise false.
 */
export function checkImage(mime) {
  if (mime === "") return true;
  const allowedTypes = images().concat(archives());
  return allowedTypes.includes(mime);
}

/**
 * Checks if the given MIME type is allowed for intros, demos and cracktros.
 *
 * @param {string} mime - The MIME type to check.
 * @returns {boolean} - Returns true if the MIME type is allowed, otherwise false.
 */
export function checkIntro(mime) {
  if (mime === "") return true;
  const allowedTypes = apps().concat(archives(), binaries());
  return allowedTypes.includes(mime);
}

/**
 * Checks if the given MIME type is allowed for magazines and newletters.
 *
 * @param {string} mime - The MIME type to check.
 * @returns {boolean} - Returns true if the MIME type is allowed, otherwise false.
 */
export function checkMagazine(mime) {
  if (mime === "") return true;
  const allowedTypes = texts().concat(archives(), apps(), binaries());
  return allowedTypes.includes(mime);
}

/**
 * Checks if the given MIME type is allowed for text files.
 *
 * @param {string} mime - The MIME type to check.
 * @returns {boolean} - Returns true if the MIME type is allowed for text files, otherwise false.
 */
export function checkText(mime) {
  if (mime === "") return true;
  const allowedTypes = texts().concat(archives());
  return allowedTypes.includes(mime);
}
