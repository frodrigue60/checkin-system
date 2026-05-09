/**
 * LocationService.ts
 * Provides utility functions for reverse geocoding and robust location retrieval.
 */

export interface GeocodeResult {
    display_name: string;
    address: {
        road?: string;
        house_number?: string;
        suburb?: string;
        city?: string;
        state?: string;
        postcode?: string;
        country?: string;
    };
}

export interface Coordinates {
    latitude: number;
    longitude: number;
    accuracy?: number;
    source: 'gps' | 'network' | 'ip';
}

/**
 * Gets the current location using a layered approach:
 * 1. High Accuracy GPS
 * 2. Network-based location
 * 3. IP-based location (Fallback)
 */
export async function getCurrentLocation(): Promise<Coordinates> {
    // Try Browser Geolocation API first
    if ('geolocation' in navigator) {
        try {
            const pos = await new Promise<GeolocationPosition>((resolve, reject) => {
                navigator.geolocation.getCurrentPosition(resolve, reject, {
                    enableHighAccuracy: true,
                    timeout: 8000,
                    maximumAge: 0
                });
            });
            return {
                latitude: pos.coords.latitude,
                longitude: pos.coords.longitude,
                accuracy: pos.coords.accuracy,
                source: 'gps'
            };
        } catch (error) {
            console.warn('High accuracy GPS failed, trying network...', error);
            
            try {
                const pos = await new Promise<GeolocationPosition>((resolve, reject) => {
                    navigator.geolocation.getCurrentPosition(resolve, reject, {
                        enableHighAccuracy: false,
                        timeout: 5000
                    });
                });
                return {
                    latitude: pos.coords.latitude,
                    longitude: pos.coords.longitude,
                    accuracy: pos.coords.accuracy,
                    source: 'network'
                };
            } catch (err) {
                console.warn('Network location failed, falling back to IP...', err);
            }
        }
    }

    // Layer 3: IP-based fallback
    try {
        const response = await fetch('https://freeipapi.com/api/json');
        if (!response.ok) throw new Error('IP Location service failed');
        const data = await response.json();
        return {
            latitude: data.latitude,
            longitude: data.longitude,
            source: 'ip'
        };
    } catch (error) {
        console.error('All location methods failed:', error);
        throw new Error('No se pudo determinar la ubicación por ningún método.');
    }
}

export async function reverseGeocode(lat: number, lon: number): Promise<string> {
    try {
        const response = await fetch(
            `https://nominatim.openstreetmap.org/reverse?lat=${lat}&lon=${lon}&format=json`,
            {
                headers: {
                    'Accept-Language': 'es',
                    'User-Agent': 'JGC-Attendance-System/1.0'
                }
            }
        );

        if (!response.ok) {
            throw new Error('Geocoding service unavailable');
        }

        const data: GeocodeResult = await response.json();
        return data.display_name || 'Ubicación desconocida';
    } catch (error) {
        console.error('Reverse geocoding error:', error);
        return 'Error al obtener dirección';
    }
}
