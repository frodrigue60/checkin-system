import { register, init, getLocaleFromNavigator } from 'svelte-i18n';

register('es', () => import('./locales/es.json'));
register('en', () => import('./locales/en.json'));

const defaultLocale = 'es';
const supportedLocales = ['es', 'en'];

function getInitialLocale() {
	if (typeof localStorage !== 'undefined') {
		const savedLocale = localStorage.getItem('locale');
		if (savedLocale && supportedLocales.includes(savedLocale)) {
			return savedLocale;
		}
	}

	const navigatorLocale = getLocaleFromNavigator()?.split('-')[0];
	if (navigatorLocale && supportedLocales.includes(navigatorLocale)) {
		return navigatorLocale;
	}

	return defaultLocale;
}

init({
	fallbackLocale: defaultLocale,
	initialLocale: getInitialLocale()
});
