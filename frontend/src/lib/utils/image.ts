/**
 * Utility to compress and resize images on the client side using Canvas API.
 */

export interface CompressionOptions {
	maxWidth?: number;
	maxHeight?: number;
	quality?: number;
	mimeType?: string;
}

/**
 * Compresses an image file and returns a Blob.
 * @param file The original File object from an input[type=file]
 * @param options Compression settings
 */
export async function compressImage(
	file: File,
	options: CompressionOptions = {}
): Promise<Blob> {
	const {
		maxWidth = 1200,
		maxHeight = 1200,
		quality = 0.8,
		mimeType = 'image/jpeg'
	} = options;

	return new Promise((resolve, reject) => {
		const reader = new FileReader();
		reader.readAsDataURL(file);
		reader.onload = (event) => {
			const img = new Image();
			img.src = event.target?.result as string;
			img.onload = () => {
				const canvas = document.createElement('canvas');
				let width = img.width;
				let height = img.height;

				// Calculate new dimensions
				if (width > height) {
					if (width > maxWidth) {
						height *= maxWidth / width;
						width = maxWidth;
					}
				} else {
					if (height > maxHeight) {
						width *= maxHeight / height;
						height = maxHeight;
					}
				}

				canvas.width = width;
				canvas.height = height;

				const ctx = canvas.getContext('2d');
				if (!ctx) {
					reject(new Error('Could not get canvas context'));
					return;
				}

				// Draw and compress
				ctx.drawImage(img, 0, 0, width, height);

				canvas.toBlob(
					(blob) => {
						if (blob) {
							resolve(blob);
						} else {
							reject(new Error('Canvas toBlob failed'));
						}
					},
					mimeType,
					quality
				);
			};
			img.onerror = (err) => reject(err);
		};
		reader.onerror = (err) => reject(err);
	});
}

/**
 * Converts a Blob to a File object.
 * @param blob The compressed blob
 * @param originalName The original filename to preserve extension/name
 * @param mimeType The target mime type
 */
export function blobToFile(blob: Blob, originalName: string, mimeType: string = 'image/jpeg'): File {
	const name = originalName.replace(/\.[^/.]+$/, "") + (mimeType === 'image/jpeg' ? '.jpg' : '.png');
	return new File([blob], name, { type: mimeType });
}
