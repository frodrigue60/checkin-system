import { get } from 'svelte/store';
import { _ } from 'svelte-i18n';

/**
 * Maps common backend error strings to i18n keys
 */
const ERROR_MAP: Record<string, string> = {
    // Auth Errors
    "Invalid email or password": "errors.invalid_credentials",
    "Email already exists or database error": "errors.email_exists",
    
    // Attendance Errors (ES)
    "Acceso denegado: Tu cuenta de empleado está inactiva. Contacta a un administrador.": "errors.account_inactive",
    "Cuenta inactiva": "errors.account_inactive",
    "Perfil incompleto: Debes tener una sede y un puesto asignados para marcar asistencia.": "errors.profile_incomplete",
    "Perfil incompleto": "errors.profile_incomplete",
    "Ya has marcado entrada hoy": "errors.already_checked_in",
    "Hoy no es un día laboral programado para tu turno.": "errors.not_work_day",
    "No se encontró una sesión activa para hoy": "errors.no_active_session",
    
    // Generic
    "Access denied": "errors.access_denied",
    "Internal server error": "errors.server_error"
};

/**
 * Translates a backend error string if it matches a known pattern,
 * otherwise returns the original string.
 */
export function translateError(rawError: string): string {
    const t = get(_);
    const key = ERROR_MAP[rawError];
    
    if (key) {
        return t(key);
    }
    
    return rawError;
}
