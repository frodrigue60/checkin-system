-- Down: migrate_attendance_to_timestamptz

-- 1. Crear columnas temporales con el formato antiguo
ALTER TABLE public.attendances ADD COLUMN date date;
ALTER TABLE public.attendances ADD COLUMN check_in_old TIME(0) WITHOUT TIME ZONE;
ALTER TABLE public.attendances ADD COLUMN lunch_start_old TIME(0) WITHOUT TIME ZONE;
ALTER TABLE public.attendances ADD COLUMN lunch_end_old TIME(0) WITHOUT TIME ZONE;
ALTER TABLE public.attendances ADD COLUMN check_out_old TIME(0) WITHOUT TIME ZONE;

-- 2. Revertir los datos extrayendo fecha y hora por separado
UPDATE public.attendances SET 
    date = CAST(check_in AS DATE),
    check_in_old = CAST(check_in AS TIME),
    lunch_start_old = CAST(lunch_start AS TIME),
    lunch_end_old = CAST(lunch_end AS TIME),
    check_out_old = CAST(check_out AS TIME);

-- 3. Restaurar nombres de columnas
ALTER TABLE public.attendances DROP COLUMN check_in;
ALTER TABLE public.attendances DROP COLUMN lunch_start;
ALTER TABLE public.attendances DROP COLUMN lunch_end;
ALTER TABLE public.attendances DROP COLUMN check_out;

ALTER TABLE public.attendances RENAME COLUMN check_in_old TO check_in;
ALTER TABLE public.attendances RENAME COLUMN lunch_start_old TO lunch_start;
ALTER TABLE public.attendances RENAME COLUMN lunch_end_old TO lunch_end;
ALTER TABLE public.attendances RENAME COLUMN check_out_old TO check_out;

ALTER TABLE public.attendances ALTER COLUMN date SET NOT NULL;
