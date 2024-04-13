// uploader.mjs

import { getElmById } from "./helper.mjs";

/**
 * Retrieves a modal object by its element ID.
 *
 * @param {string} elementId - The ID of the element representing the modal.
 * @returns {Object} - The modal object.
 * @throws {Error} - If the bootstrap object is undefined.
 */
export function getModalById(elementId) {
  if (bootstrap === undefined) {
    throw new Error(`The bootstrap object is undefined.`);
  }
  const element = getElmById(elementId);
  const modal = new bootstrap.Modal(element);
  return modal;
}

/**
 * Focuses the modal by its element ID and submission ID.
 * @param {string} elementId - The ID of the modal element.
 * @param {string} submissionId - The ID of the submission element.
 * @returns {Object} - The modal object.
 * @throws {Error} - If the submission element is null or if the bootstrap object is undefined.
 */
export function focusModalById(elementId, submissionId) {
  const input = document.getElementById(submissionId);
  if (input == null) {
    throw new Error(`The ${submissionId} element is null.`);
  }
  const element = getElmById(elementId);
  element.addEventListener("shown.bs.modal", function () {
    input.focus();
  });
  if (bootstrap === undefined) {
    throw new Error(`The bootstrap object is undefined.`);
  }
  const modal = new bootstrap.Modal(element);
  return modal;
}

/**
 * Checks the SHA-384 hash of a file by sending it to the server.
 * This function is a client convenience to save time and bandwidth.
 * If the client browser does not support the required APIs,
 * it does not matter as the hash is rechecked on the server after uploading.
 * @param {File} file - The file to be hashed.
 * @returns {Promise<boolean>} - A promise that resolves to true if the server confirms the hash, false otherwise.
 */
export async function checkSHA(file) {
  try {
    const hash = await sha384(file);
    const response = await fetch(`/uploader/sha384/${hash}`, {
      method: "PUT",
      headers: {
        "Content-Type": "text/plain",
      },
      body: hash,
    });
    if (!response.ok) {
      throw new Error(
        `Hashing is not possible, server response: ${response.status}`
      );
    }
    const responseText = await response.text();
    return responseText == "true";
  } catch (e) {
    console.log(`Hashing is not possible: ${e}`);
  }
}

async function sha384(file) {
  try {
    const buffer = await file.arrayBuffer();
    const hashBuffer = await crypto.subtle.digest("SHA-384", buffer);
    return Array.from(new Uint8Array(hashBuffer))
      .map((b) => b.toString(16).padStart(2, "0"))
      .join("");
  } catch (e) {
    throw new Error(`Could not use arrayBuffer or crypto.subtle: ${e}`);
  }
}
