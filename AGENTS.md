# Attendance Management System - AGENTS & DEVELOPER CONTEXT

## Project Description

This project is an advanced, enterprise-grade Employee Attendance and Workforce Management System. Designed to be scalable and adaptable to diverse organizational structures, it enables precise tracking of Check-In, Lunch, and Check-Out times while enforcing strict geolocation perimeter validation and configurable Work Shift rules. The core engine features automated incident detection (such as Late arrivals and Out of Range events) alongside dynamic financial reporting that supports automated compliance data and penalty deductions.

---

## Architecture & Core Entities

The system relies on a relational database structure backed strictly by **PostgreSQL**.

### 1. **User**
- **Role**: Authentication & Authorization wrapper.
- **Relations**: Linked to a specific access role and optionally associated with an operational Employee profile.
- **Key Attributes**: Role Identifier.

### 2. **Employee**
- **Role**: The central entity representing a staff member within the workforce.
- **Relations**:
    - Linked to authentication credentials.
    - Designated to a primary operational location (Work Center).
    - Assigned to a scheduled Work Shift.
    - Associated with a Job Position (determining financial parameters).
    - Connected to historical Attendance tracking logs.
- **Key Attributes**: Operational Status (Active/Inactive), Inherited Financial Rates.

### 3. **Attendance** (Snapshot Entity)
- **Role**: Validated daily record of a worker's activity.
- **Relations**:
    - Belongs to an Employee.
    - **[NEW]** Captures the exact Work Center location _at the moment of the initial Check-In_.
- **Key Attributes**:
    - **Unified Temporal Schema**: All event markers (`check_in`, `lunch_start`, `lunch_end`, `check_out`) utilize `TIMESTAMPTZ`. This ensures precision across calendar boundaries (near-midnight shifts) and prevents legacy year-0 arithmetic bugs.
    - **Evidence Validation**: The `check_out` event requires a valid `evidence_url` (captured image) to ensure service completion and auditability.
    - Literal GPS Snapshots (Latitude, Longitude).
    - Session Totals (Net hours worked, Daily earnings).
- **Business Logic**:
    - **Multi-Session Support**: For "Field" shifts, employees can perform multiple check-in/out cycles per day. The system tracks the "active" session and allows reopening a new one once the previous is closed.
    - **Snapshotting**: The location context is frozen upon creation. This guarantees historical financial integrity even if an employee is functionally transferred to a new perimeter later in the fiscal period.
    - **Robust Duration Calc**: Hours are computed via native PostgreSQL interval arithmetic on unified timestamps, ensuring that `check_out` always follows `check_in` chronologically within the same logical tracking session.

---

## Technical Debt & Robustness (April 2026 Refactor)

### 1. Unified Timestamp Unification
The system was migrated from split `DATE` and `TIME` columns to unified `TIMESTAMPTZ`.
- **Reasoning**: Legacy split columns led to "Year 0000" scanning errors in Go and multi-million hour duration calculation overflows when cross-referencing calendar dates.
- **Implementation**: Native timezone support preserves the "Moment of Truth" regardless of server or database local settings.

### 2. Incident Detection Refinement
Incident detection (Late, Out of Range) now occurs against the unified `check_in` reference.
- **Late Arrivals**: Evaluated by comparing the time components of the `check_in` timestamp against the assigned `work_shift` expected start, adjusted for the designated grace period.
- **Out of Range**: Evaluated during every event lifecycle (Check-in, Lunch Start/End, Check-out).

### 4. **Work Center**
- **Role**: Physical geofenced location or branch where operations occur.
- **Relations**:
    - **[NEW]** Overseen by a designated management profile.
    - Hosts multiple assigned staff members.
- **Key Attributes**: Mathematical Geolocation (Latitude, Longitude), Configurable Tolerance Radius (expressed in meters).

### 4.5. **Work Shift**
- **Role**: Temporal rules defining expected behavior.
- **Types**:
    - `fixed`: Standard geofence and strict timing. Generates `late` and `out_of_range` incidents.
    - `flexible`: No `late` incidents. Focuses only on duration.
    - `field`: Mobile workforce. Disables `late` AND `out_of_range` incidents to allow movement across express requests.

### 5. **Position**
- **Role**: Job definition catalog dictating financial computation rules.
- **Key Attributes**:
    - Standard base pay rate.
    - **[NEW]** Fixed deduction penalty for tardiness (Late).
    - **[NEW]** Fixed deduction penalty for location violations (Out of Range).

### 6. **Report**
- **Role**: Aggregated financial and operational statement per employee representing a specific payroll sequence.
- **Key Attributes**: Aggregate hours, Net total earnings, Absolute incidents count, Granular Daily Breakdown.
- **Daily Breakdown Context**: Contains procedural day-by-day dataset including Date, Computable Hours, Standard Earnings, Deductions applied, Shift completion status, and the human-readable Location Name.

### 7. **Incident**
- **Role**: Autosaved infractions detected autonomously during routine checks.
- **Types**: Tardiness (`late`), GPS proximity failures (`out_of_range`).
- **Relations**: Cross-linked to the offending daily tracking log and the evaluated operational perimeter.

---

## Key Workflows & Business Logic

### 1. Operational Check-In Process

1.  **Validation Phase**:
    - Global Calendar: Is the current date flagged as a **Mandatory Holiday**? (Prevents standard operations).
    - Profile Status: Is the requesting agent active within the corporate directory?
    - Geographic Proximity: Evaluated for all events. Out-of-range events are permitted to proceed but automatically trigger an `out_of_range` incident for administrative review and potential deduction.
2.  **Execution Phase**:
    - Instantiates a daily operational tracking log.
    - **Context Capture**: Permanently saves the current location reference and literal geospatial coordinates.
3.  **Anomaly Detection (Incident Generation)**:
    - Timing Logic: If `Action Timestamp > Expected Shift Time + Grace Period Tolerance` -> Spawns a `Late` Incident marker.
    - Distance Logic: If `Calculated Distance > Perimeter Tolerance Radius` -> Spawns an `Out of Range` Incident marker.
    - Lunch Logic: If `Lunch Duration > Allowed Lunch Time` -> Spawns a `Lunch Overstay` Incident marker.

### 2. Analytics & Financial Report Generation

1.  **Data Isolation**: Queries historical logs within a requested temporal boundary per target employee.
2.  **Computational Engine**:
    - Aggregates strictly net active hours and proportional gross earnings.
    - **Penalty Engine**: Iterates through detected infractions. Only incidents with status `pending` or `approved` result in financial deductions. `Justified` incidents do not apply penalties. For instance, finding a `Late` marker subtracts the penalty constant defined by the employee's active Job Position.
3.  **Persistence**: Finalizes and archives the temporal snapshot into the centralized ledger.
4.  **Presentation Rules**:
    - Ledgers showing financial penalties must trigger visual warnings.
    - Incompletely closed operational days (missing check-out logs) must raise administrative alerts.

---

## 🤖 LLM & Systemic Guardrails (Context for AI Agents)

To maintain the architectural integrity of the JGC Check-in system, future modifications MUST adhere to the following constraints:

### 1. **Temporal & Database Invariants**
- **Strict TIMESTAMPTZ**: All time-sensitive events (`check_in`, `lunch_start`, `lunch_end`, `check_out`) MUST use `TIMESTAMPTZ`. Never use separate `DATE` and `TIME` columns.
- **Duration Arithmetic**: Perform duration calculations using PostgreSQL native arithmetic (e.g., `check_out - check_in`) or Go `time.Duration` on unified timestamps.
- **Nullable Fields**: `NetHoursWorked` and `DailyEarnings` are pointers (`*float64`) in models to handle unclosed sessions safely. Always check for `nil` before dereferencing.
- **Daily Filtering**: When querying for "today's" record, use the date cast: `WHERE check_in::date = $1`.

### 2. **Business Logic Constraints**
- **Idempotency**: An employee can have exactly ONE active attendance record per logical work session.
- **Hierarchy of events**: `check_in` < `lunch_start` < `lunch_end` < `check_out`. Validations must verify this chronology.
- **Geofencing**: Every state change (Check-In, Lunch Start, etc.) MUST independently calculate the distance from the `WorkCenter` coordinates and compare against `tolerance_radius_meter`.

### 3. **Common Patterns (Contextual Anchors)**
- **SOLID Handlers**: Avoid "God Object" handlers. Handlers MUST be domain-specific (e.g., `ShiftHandler`, `CenterHandler`) and embed `AdminBase` (Facade Pattern) for shared dependencies.
- **Querier Interface**: Services MUST depend on `database.Querier` rather than concrete `*sqlx.DB` or `*sqlx.Tx`. This allows the same service method to be used safely in both isolated queries and transactional blocks.
- **Scanning**: Use `sqlx` and ensure struct tags match exact database column names.
- **Hander Errors**: Standardize internal errors as `500 Internal Server Error` and business validation failures as `400 Bad Request` or `403 Forbidden`.
- **Mappers**: Always use `MapAttendanceToDTO` (and similar mappers) to sanitize sensitive fields and format dates for the frontend.

### 4. **Identified Technical Debt (Maintenance Required)**
- **Report Archival**: Ensure that when an attendance record is updated after a report was generated, the report is invalidated or regenerated.
- **Midnight Shifts**: The current `::date` logic assumes single-date shifts. If night-shifts are introduced, logic must pivot to a "Work Day" anchor rather than a calendar date cast.

### 5. **Quality & Testing Guardrails**
- **Unit Testing Mandatory**: Any modification to the `AttendanceService` or `ReportService` MUST be accompanied by unit tests.
- **Mocking**: Use `sqlmock` to simulate database interactions in services. Avoid using a real database for unit tests to ensure speed and isolation.
- **Coverage**: Maintain high coverage for core arithmetic logic (earnings, durations, distance).

### 6. **Frontend Type Safety & Audit**
- **Strict Typing**: Never use `any` for data models. Always use the interfaces defined in `lib/types/models.ts`.
- **Generic API Calls**: Use `apiFetch<T>` to ensure responses are typed from the moment of retrieval.
- **Centralized Auditing**: All API failures must be logged via the centralized console system in `api.ts` to facilitate technical auditing.
- **Audit Logs**: Any modification to critical entities MUST be reflected in the `admin/audit` module, ensuring transparency in administrative actions.

### 8. **Infrastructure & Performance Standards**
- **Connection Pooling**: Always ensure `DB.SetMaxOpenConns` and `DB.SetMaxIdleConns` are configured. Never leave default connection limits in production.
- **Context-Aware Queries**: Use `DB.SelectContext` or `DB.GetContext` with `context.Context` (cascaded directly from Fiber's `c.Context()`) for all HTTP-triggered queries to support request cancellation. Background workers MUST use `context.WithTimeout`.
- **N+1 Prevention**: Avoid querying the database inside loops. Use `sqlx.In` or JOINs to fetch related data in a single round-trip.
- **Structured Logging**: The use of `fmt.Printf` and `fmt.Println` is BANNED in production code. Always use `utils.GetLogger()` with structured `zap` fields for reliable observability.

### 9. **Business Logic & Validation Standards**
- **Centralized Validation**: Use `utils.ParseAndValidate(c, &req)` for all POST/PUT requests to ensure type safety and constraint enforcement via `go-playground/validator`.
- **Service-Driven Logic**: Complex logic (Earnings, Distance, Incident detection) MUST reside in their respective Services (e.g., `AttendanceService`).
- **Transactional Consistency (Unit of Work)**: Destructive or multi-table operations (e.g., UpdateAttendance, UpdateIncidentStatus) MUST be wrapped in a database transaction (`DB.BeginTxx`). Pass the transaction down to the service layer using the `database.Querier` interface to guarantee atomic commits or rollbacks.
- **Flexible Time Parsing**: Use `s.ParseFlexibleTime` to support varied timestamp formats and prevent parsing errors.

### 7. **Security & Error Handling Standards**
- **JWT Algorithm Validation**: All JWT validations MUST explicitly check for the expected signing method (e.g., `*jwt.SigningMethodHMAC`). Never rely on the default `Parse` without algorithm verification.
- **Error Masking**: Internal system errors (SQL, File System, etc.) MUST NEVER be sent directly to the client. Use `utils.SendError(c, code, userFriendlyMessage, realError)` to log the detail internally while protecting the API surface.
- **Restrictive CORS**: The `AllowOrigins` policy MUST be specific and loaded from environment variables. Avoid using `*` in production environments.
- **Swagger Gating**: Documentation endpoints MUST be gated by an environment variable (e.g., `ENABLE_SWAGGER`) to prevent unauthorized reconnaissance in production.

### 10. **Timezone-Aware Filtering**
- **The Problem**: Filtering by a calendar date using `::date` on UTC timestamps miscategorizes records that occurred during the "local night" (e.g., a record at 11:30 PM local might be 05:30 AM UTC next day).
- **Mandatory Pattern**: Calculate day boundaries (`00:00:00` and `23:59:59`) in the client's local timezone and send them as ISO UTC strings to the backend.
- **Backend Implementation**: Use precise timestamp comparisons (`>=` and `<=`) instead of date casting to ensure the "window of time" matches the user's local day perfectly.

### 12. **Mobile Workforce & Evidence Standards**
- **Validation**: Any `check_out` request MUST include `evidence_url`. If null, the request MUST fail with 400.
- **Field Shift Logic**: When evaluating incidents for `field` shifts, the distance check should be bypassed to accommodate "express" mobile requests.

### 11. **JSON Collection Consistency**
- **Rule**: All list and paginated endpoints MUST return an empty array `[]` instead of `null` when no data is found.
- **Go Pattern**: Initialize slices using literals: `logs := []models.AuditLog{}` instead of `var logs []models.AuditLog`. This ensures the JSON encoder produces `[]` instead of `null`.

### 13. **Temporal Integrity & Legacy Handling**
- **ISO 8601 UTC Standard**: ALL date and time data transmitted via API MUST use ISO 8601 strings in UTC format (e.g., `2026-05-09T14:30:00Z`).
- **Legacy "Year 0" Awareness**: 
    - **Backend**: Go's zero-time (`0001-01-01`) or SQL legacy dates (`0000-01-01`) are often used as placeholders for "Time-only" fields (like `WorkShift` start/end).
    - **Frontend Parsing**: Never discard a timestamp just because it contains `0000` or `0001` years. Always extract the time component (`HH:mm`) unless the field is truly optional/null.
- **Time-Only Extraction Rule**: When rendering shift schedules, ignore the date component. Use a robust helper like `formatTime` that gracefully handles `0001-01-01` by returning the time part while treating truly empty values as `--:--`.
- **Arithmetic Safety**: 
    - Always use PostgreSQL native interval arithmetic for durations (`check_out - check_in`).
    - In Svelte, use dedicated helper functions (`timeToMinutes`, `calculateDuration`) to handle midnight-crossing shifts by adding `24 * 60` minutes when the end time is numerically lower than the start time.

---

## 🎨 Frontend & UI UX Guidelines

### 1. **Transition Persistence (Svelte 5)**
- **The Problem**: In SvelteKit, `in:fly` transitions are often skipped on the initial page load (SSR/Hydration) because the elements already exist in the static HTML.
- **Mandatory Pattern**: To ensure "Premium" entrance animations always play, wrap the main page content in a `{#if mounted}` block.
- **Implementation**:
```svelte
<script>
  let mounted = $state(false);
  onMount(() => mounted = true);
</script>

{#if mounted}
  <main in:fly={{ y: 20, duration: 800 }}>
    ...
  </main>
{/if}
```
- **Rationale**: This guarantees that the high-fidelity "vuelo" effect is visible regardless of whether the user performed an internal navigation or a direct URL entry.

### 2. **Aesthetic Consistency**
- **Skeletons vs Loaders**: Use `animate-pulse` skeletons for initial data fetching to maintain layout stability. Use `Loader2` (animate-spin) for action-triggered states (e.g., saving, generating reports).
- **Transitions**: Use `animate:flip` on list items and `in:fly` on the container or individual entries to ensure a responsive, editorial feel.

### 3. **Event Bubbling & Interactive Cards**
- **The Problem**: Wrapping an entire card in an `<a>` tag interferes with internal interactive elements like checkboxes or delete buttons.
- **Mandatory Pattern**: Do NOT wrap the entire card container in an `<a>` tag. Move the navigation link specifically to the entity name or a "View" button.
- **Checkbox Safety**: Ensure checkboxes in lists have dedicated containers and do not trigger parent link events.
