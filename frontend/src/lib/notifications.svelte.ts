export type NotificationType = 'success' | 'error' | 'info' | 'warning';

export interface Notification {
	id: string;
	type: NotificationType;
	message: string;
	duration: number;
	createdAt: number;
}

class NotificationStore {
	#notifications = $state<Notification[]>([]);

	get items() {
		return this.#notifications;
	}

	add(message: string, type: NotificationType = 'info', duration: number = 5000) {
		// Cap notifications to prevent memory issues
		if (this.#notifications.length > 5) {
			this.#notifications.shift();
		}

		const id = crypto.randomUUID();
		const notification: Notification = {
			id,
			type,
			message,
			duration,
			createdAt: Date.now()
		};

		this.#notifications.push(notification);

		if (duration > 0) {
			setTimeout(() => {
				this.dismiss(id);
			}, duration);
		}

		return id;
	}

	success(message: string, duration?: number) {
		return this.add(message, 'success', duration);
	}

	error(message: string, duration: number = 8000) {
		return this.add(message, 'error', duration);
	}

	info(message: string, duration?: number) {
		return this.add(message, 'info', duration);
	}

	warning(message: string, duration?: number) {
		return this.add(message, 'warning', duration);
	}

	dismiss(id: string) {
		const index = this.#notifications.findIndex((n) => n.id === id);
		if (index !== -1) {
			this.#notifications.splice(index, 1);
		}
	}
}

export const notifications = new NotificationStore();
