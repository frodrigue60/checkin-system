import { browser } from '$app/environment';

import { type User, type RoleSlug } from './types/models';

class AuthState {
	token = $state<string | null>(null);
	user = $state<User | null>(null);

	constructor() {
		if (browser) {
			this.token = localStorage.getItem('jwt_token');
			const u = localStorage.getItem('user_data');
			if (u && u !== 'undefined') {
				try {
					this.user = JSON.parse(u);
				} catch (e) {
					console.error('Failed to parse user data from storage', e);
				}
			}
		}
	}

	login(token: string, user: User) {
		this.token = token;
		this.user = user;
		if (browser) {
			localStorage.setItem('jwt_token', token);
			localStorage.setItem('user_data', JSON.stringify(user));
		}
	}

	logout() {
		this.token = null;
		this.user = null;
		if (browser) {
			localStorage.removeItem('jwt_token');
			localStorage.removeItem('user_data');
		}
	}

	get isAuthenticated() {
		return !!this.token;
	}

	get isAdmin() {
		return this.user?.role_slug === 'admin';
	}

	get isManager() {
		return this.user?.role_slug === 'manager';
	}

	get isEmployee() {
		return this.user?.role_slug === 'employee';
	}

	get isSupervisor() {
		return this.user?.role_slug === 'supervisor';
	}

	get isBaseUser() {
		return this.user?.role_slug === 'user';
	}
}

export const authState = new AuthState();
