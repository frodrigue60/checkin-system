# 🗺️ Roadmap de Migración y Desarrollo: JGC Check-in

Este documento detalla el progreso actual de la migración tecnológica desde el monolito Laravel hacia una arquitectura desacoplada moderna: **Backend en Go (Fiber)** y **Frontend en SvelteKit (Executive Ethereal UI)**.

---

## ✅ Hitos Completados (Implementado)

### 1. 🛡️ Infraestructura de Backend (Go Fiber)

- **Seguridad de Grado Empresarial:** Implementación de autenticación **Bearer JWT** pura (sin dependencias de sesión).
- **Control de Acceso Dinámico (RBAC):** Middleware `RoleCheck` adaptado para autorización por `slug_id` (admin, manager, employee).
- **Lógica de Asistencia Refinada:**
  - Soporte para **Turnos Nocturnos** (Night Shift) que cruzan la medianoche.
  - Cálculo automático de **Horas Netas** restando periodos de almuerzo.
  - Validación de **Geofencing** mediante fórmula de Haversine en el backend.
- **Módulo Administrativo (Full CRUD):**
  - `WorkCenters`: Configuración de geovallas y radios de tolerancia.
  - `WorkShifts`: Gestión de jornadas y periodos de gracia.
  - `Positions`: Estructura salarial y configuración de penalidades por retardo/ubicación.
  - `Employees`: Administración de contratos, roles y estados.
  - `Holidays`: Bloqueo de operaciones en días feriados legales.
- **Motor de Reportes (Fase 5 Completada):**
  - Generador transaccional en runtime (evita inconsistencias de datos).
  - Endpoints de histórico y desglose detallado por rango de fechas.
  - **Filtrado Avanzado:** Capacidad de generar reportes segmentados por Sede, Puesto y Turno.
  - **Exportación PDF:** Generación de documentos corporativos inmutables con saltos de página automáticos por empleado.
- **Seguridad & Hardening (Fase 1 Audit):**
  - **Validación Estricta JWT:** Validación explícita de algoritmos (HMAC) para prevenir ataques de suplantación.
  - **Error Masking:** Implementación de helper `SendError` para enmascarar detalles internos de DB y servidor.
  - **CORS Restrictivo:** Migración de política "allow-all" a lista blanca configurable por entorno.
  - **Swagger Gating:** Exposición condicional de la documentación técnica mediante flags de entorno.
- **Optimización de Infraestructura (Fase 2 Audit):**
  - **Connection Pooling:** Configuración de límites de conexiones simultáneas y tiempo de reciclaje.
  - **Eliminación de N+1:** Refactorización de endpoints críticos para reducir latencia mediante carga masiva de datos.
  - **Context Timeouts:** Implementación de límites de tiempo de ejecución para evitar bloqueo de hilos por consultas lentas.
- **Robustez de Lógica & Validación (Fase 3 & 4 Audit):**
  - **Centralización de Lógica de Negocio:** Migración de cálculos financieros e incidentes a `AttendanceService`.
  - **Validación Estructurada:** Implementación de `go-playground/validator` y helper `ParseAndValidate` en todos los flujos de entrada.
  - **Consistencia Transaccional:** Uso de `AutoDetectIncidentsTx` para sincronizar estados de asistencia y cálculos económicos en una sola operación atómica.
  - **Refactorización de Handlers:** Simplificación de controladores de asistencia delegando la detección de geofencing y sobretiempos al servicio central.

### 2. ✨ Experiencia de Usuario (SvelteKit)

- **Diseño "Executive Ethereal":** Implementación de una estética premium basada en Glassmorphism, tonal layering y tipografía editorial (Inter/Manrope).
- **Arquitectura Reactiva:** Gestión de estado global mediante **Svelte Runes ($state)** para autenticación y persistencia de UI.
- **Sistema de Componentes Premium:**
  - `Table.svelte`: Tablas interactivas con búsqueda, filtrado y estados de carga.
  - `Modal.svelte`: Diálogos animados y accesibles para flujos de CRUD.
- **Management Hub:** Panel administrativo con estadísticas clave (kpis) y navegación rápida hacia módulos operativos.
- **Paneles Administrativos Dinámicos:** Vistas completas para la gestión de todos los catálogos del sistema con feedback inmediato.
- **🌍 Internacionalización (i18n):** Configuración de `svelte-i18n` con carga asíncrona y selector global en el `Sidebar`.

---

## 🌍 Roadmap de Internacionalización (i18n)

Plan de fases para la implementación total de soporte multi-idioma.

### Fase 1: Infraestructura y Núcleo (Completado)

- [x] Configuración de `svelte-i18n` y carga asíncrona.
- [x] Selector de idioma reactivo en Sidebar.
- [x] Internacionalización de componentes base (`Table`, `MobileBottomNav`, `AbsenceModal`).
- [x] Traducción completa del Dashboard del Empleado (`+page.svelte`).

### Fase 2: Autenticación y Onboarding (Completado)

- [x] **Vistas de Acceso:** Refactorización de `/login` y `/register`.
- [x] **Feedback de Sistema:** Localización de mensajes de validación y errores de API.
- [ ] **Detección Automática:** Implementación de detección de idioma del navegador.

### Fase 3: Portal del Empleado (ESS) (Completado)

- [x] **Historial y Horarios:** Localización de `/history` y `/schedule`.
- [x] **Perfil:** Traducción de la configuración de cuenta en `/profile`.
- [x] **Incidencias:** Traducción dinámica de tipos de incidencia (Retardos, Fuera de Rango).

### Fase 4: Administración Operativa (Completado)

- [x] **Gestión de Sedes:** Localización del módulo de Centros de Trabajo.
- [x] **Turnos y Puestos:** Traducción de interfaces de configuración operativa.
- [x] **Formularios CRUD:** Refactorización de todos los modales (labels, placeholders).

### Fase 5: Reportes e Inteligencia (Completado)

- [x] **Módulo de Reportes:** Localización de filtros y visualización de nóminas.
- [x] **PDF i18n:** Soporte de traducción para documentos exportables (Go Backend).
- [x] **Formatos Locales:** Adaptación de moneda y fechas según el locale (`Intl` API).

### Fase 6: Refinamientos Finales

- [x] **Detección Automática:** Implementación de detección de idioma del navegador y persistencia local.
- [/] **Mapeo de Errores:** Implementación parcial mediante `error-translator.ts` en el frontend; pendiente migrar el backend a códigos de error inmutables.

---

## 🚧 En Desarrollo Activo (Fase 5: Reportes y Refinamiento)

### 🛠️ Próximas Implementaciones Inmediatas

- [x] **Exportación nativa a PDF (Go Backend):** Generación de documentos corporativos inmutables usando `maroto/v2`.
- [ ] **Widget de Check-in (UX Empleado):** Finalización del panel de empleado con geolocalización interactiva y estado de jornada en vivo.
- [x] **DTO & Types (SvelteKit):** Refuerzo del tipado mediante interfaces TypeScript que repliquen estrictamente los Structs de Go.
- [x] **Batch Actions:** Selección múltiple en tablas para activaciones/desactivaciones masivas de personal o centros.

### Fase 9: Evolución de Fuerza Laboral Móvil (Field Service) (Completado)

- [x] **Soporte Multisesión:** Capacidad de realizar múltiples entradas/salidas por día para personal dinámico.
- [x] **Tipos de Turno Evolucionados:** Implementación de turnos `flexible` y `field` (sin penalizaciones de geocerca/tiempo).
- [x] **Validación de Evidencia:** Implementación obligatoria de captura fotográfica para el cierre de jornada.
- [x] **Filtros Administrativos:** Expansión de búsqueda por Tipo de Turno y Puesto en todos los módulos de auditoría.

---

## 🚀 Próximas Fases (Backlog Estratégico)

### Fase 6: Optimización de Rendimiento

- [x] **Caché en Memoria:** Implementación de `go-cache` en `AdminHandler` para Centros, Turnos y Puestos (Reducción de latencia en catálogos).
- [ ] **Websockets / SSE:** Notificaciones en tiempo real para administradores ante incidencias de "Fuera de Rango".
- [x] **Asincronía Real:** Motor de `ReportJobs` implementado (Goroutines + Polling reactivo + Descarga directa).

### Fase 7: Analítica Avanzada

- [ ] **Dashboard Visual:** Gráficas de cumplimiento y puntualidad global (Chart.js / LayerChart).
- [x] **Logs de Auditoría:** Registro estricto de revisiones y ediciones manuales de registros de asistencia (Audit Hub).
- [x] **Justificación de Incidencias:** Flujo de aprobación para que empleados justifiquen retardos desde su panel.

### Fase 8: Integraciones Externas

- [ ] **Webhook Engine:** Sincronización de nómina con sistemas contables externos.
- [ ] **MFA:** Autenticación de dos factores para roles de administración.

---

## 📐 Notas Técnicas de la Migración

- **Backend:** Fiber v2 + Sqlx + PostgreSQL.
- **Frontend:** SvelteKit + TailwindCSS (Custom Premium System).
- **Auth:** JWT Bearer (Stored in LocalStorage).
- **Architecture:** Decoupled RESTful API.
