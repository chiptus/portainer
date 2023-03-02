export function readFileAsArrayBuffer(
  file: File,
  maxSize?: number
): Promise<string | ArrayBuffer | null> {
  return new Promise((resolve, reject) => {
    if (maxSize && file.size > maxSize) {
      const rounded = Math.round((maxSize / (1024 * 1024)) * 10) / 10; // 10 multiplier to round to 1 decimal place
      reject(new Error(`The uploaded file should be less than ${rounded}MB`));
    }
    const reader = new FileReader();
    reader.readAsArrayBuffer(file);
    reader.onload = () => {
      resolve(reader.result);
    };
    reader.onerror = (error) => reject(error);
  });
}

export function arrayBufferToBase64(buffer: ArrayBuffer) {
  let binary = '';
  const bytes = new Uint8Array(buffer);
  const len = bytes.byteLength;
  for (let i = 0; i < len; i += 1) {
    binary += String.fromCharCode(bytes[i]);
  }
  return window.btoa(binary);
}

export function readFileAsText(
  file: File,
  maxSize?: number
): Promise<string | ArrayBuffer | null> {
  return new Promise((resolve, reject) => {
    if (maxSize && file.size > maxSize) {
      const rounded = Math.round((maxSize / (1024 * 1024)) * 10) / 10; // 10 multiplier to round to 1 decimal place
      reject(new Error(`The uploaded file should be less than ${rounded}MB`));
    }
    const reader = new FileReader();
    reader.readAsText(file);
    reader.onload = () => {
      resolve(reader.result);
    };
    reader.onerror = (error) => reject(error);
  });
}
