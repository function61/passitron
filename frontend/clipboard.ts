
interface ClipboardDataInterface {
	setData(format: string, data: string): boolean;
}

declare global {
	interface Window {
		clipboardData: ClipboardDataInterface;
	}
}

// thanks https://stackoverflow.com/a/33928558
export default function(text: string): boolean {
	if (window.clipboardData && window.clipboardData.setData) {
		// IE specific code path to prevent textarea being shown while dialog is visible.
		return window.clipboardData.setData('Text', text);

	} else if (document.queryCommandSupported && document.queryCommandSupported('copy')) {
		const textarea = document.createElement('textarea');
		textarea.textContent = text;
		textarea.style.position = 'fixed';  // Prevent scrolling to bottom of page in MS Edge.
		document.body.appendChild(textarea);
		textarea.select();
		try {
			return document.execCommand('copy');  // Security exception may be thrown by some browsers.
		} catch (ex) {
			return false;
		} finally {
			document.body.removeChild(textarea);
		}
	}

	return false;
}
