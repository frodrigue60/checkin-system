export type RoleSlug = 'admin' | 'manager' | 'supervisor' | 'employee' | 'user';

export interface Role {
    id: number;
    name: string;
    slug: RoleSlug;
}

export interface User {
    id: number;
    name: string;
    email: string;
    phone?: string;
    role_id: number;
    role_slug?: RoleSlug;
    created_at?: string;
    updated_at?: string;
    employee_id?: number;
}

export interface Employee {
    id: number;
    user_id: number;
    position_id: number;
    work_center_id: number;
    work_shift_id: number;
    is_active: boolean;
    created_at?: string;
    updated_at?: string;
}

export interface WorkCenter {
    id: number;
    name: string;
    address?: string;
    latitude: number;
    longitude: number;
    tolerance_radius: number;
    manager_id?: number;
    timezone: string;
    created_at?: string;
    updated_at?: string;
}

export interface WorkShift {
    id: number;
    name: string;
    start_time: string; // "HH:MM:SS"
    end_time: string;
    lunch_duration_limit: string;
    grace_period: string;
    is_night_shift: boolean;
    is_active: boolean;
    enforce_lateness: boolean;
    enforce_lunch_limit: boolean;
    enforce_geofence: boolean;
    shift_type: 'fixed' | 'flexible' | 'field';
    work_days: number[]; // From JSON array [1,2,3,4,5]
    created_at?: string;
    updated_at?: string;
}

export interface Position {
    id: number;
    name: string;
    base_pay: number;
    late_penalty: number;
    out_of_range_penalty: number;
    lunch_overstay_penalty: number;
    employees_count: number;
    created_at?: string;
    updated_at?: string;
}

export interface Attendance {
    id: number;
    employee_id: number;
    work_shift_id?: number;
    work_center_id?: number;
    check_in?: string; // ISO string
    lunch_start?: string;
    lunch_end?: string;
    check_out?: string;
    check_in_latitude?: number;
    check_in_longitude?: number;
    check_out_latitude?: number;
    check_out_longitude?: number;
    net_hours_worked?: number;
    daily_earnings?: number;
    is_absence: boolean;
    absence_reason?: string;
    evidence_url?: string;
    check_out_note?: string;
    check_out_address?: string;
    created_at?: string;
    updated_at?: string;
}

export interface Incident {
    id: number;
    employee_id: number;
    work_center_id: number;
    attendance_id: number;
    type: 'late' | 'out_of_range';
    description?: string;
    is_late: boolean;
    delay_minutes: number;
    is_out_of_range: boolean;
    distance: number;
    status: 'pending' | 'approved' | 'justified' | 'rejected';
    resolved_by?: number;
    resolution_note?: string;
    metadata_json?: string;
    created_at?: string;
    updated_at?: string;
    justification?: Justification;
}

export interface IncidentRichDTO extends Incident {
    employee_name: string;
    work_center_name: string;
    attendance_date: string;
}

export interface Justification {
    id: number;
    incident_id: number;
    employee_id: number;
    message: string;
    evidence_url?: string;
    status: 'pending' | 'approved' | 'rejected';
    resolved_by?: number;
    resolution_note?: string;
    created_at?: string;
    updated_at?: string;
}

export interface Holiday {
    id: number;
    name: string;
    date: string;
    description?: string;
    type: 'mandatory' | 'optional';
    created_at?: string;
    updated_at?: string;
}

export interface SystemAlert {
    id: number;
    type: string;
    severity: 'info' | 'warn' | 'error' | 'success';
    message: string;
    metadata_json?: string;
    is_read: boolean;
    created_at: string;
}

export interface ReportJobDTO {
    id: number;
    status: 'pending' | 'processing' | 'completed' | 'failed';
    progress: number;
    processed_records: number;
    total_records: number;
    start_date: string;
    end_date: string;
    created_at: string;
}

export interface DailyBreakdownItem {
    date: string;
    check_in: string;
    lunch: string;
    check_out: string;
    net_hours?: string;
    earnings?: string;
    deduction: number;
    work_center_name: string;
    is_incomplete: boolean;
    is_holiday: boolean;
}

export interface AttendanceExportDTO {
    id: number;
    employee_name: string;
    center_name: string;
    position_name: string;
    check_in: string;
    check_out: string;
    hours: number;
    earnings: number;
    is_late: boolean;
    is_absence: boolean;
    absence_reason: string;
}

export interface AuditLog {
    id: number;
    user_id: number;
    user_name: string;
    action: string;
    entity_type: string;
    entity_id: number;
    old_value?: string;
    new_value?: string;
    ip_address?: string;
    user_agent?: string;
    created_at: string;
}

export type AuditAction = 
    | 'auth_login' 
    | 'auth_logout' 
    | 'create_employee' 
    | 'update_employee' 
    | 'delete_employee' 
    | 'create_center' 
    | 'update_center' 
    | 'delete_center'
    | 'create_shift'
    | 'update_shift'
    | 'delete_shift'
    | 'create_position'
    | 'update_position'
    | 'delete_position'
    | 'generate_report'
    | 'resolve_incident';
