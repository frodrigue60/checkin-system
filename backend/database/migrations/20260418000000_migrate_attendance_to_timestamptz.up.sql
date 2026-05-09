-- Up: migrate_attendance_to_timestamptz

-- 1. Agregar columnas nuevas de tipo TIMESTAMPTZ
ALTER TABLE public.attendances ADD COLUMN check_in_tz TIMESTAMPTZ;
ALTER TABLE public.attendances ADD COLUMN lunch_start_tz TIMESTAMPTZ;
ALTER TABLE public.attendances ADD COLUMN lunch_end_tz TIMESTAMPTZ;
ALTER TABLE public.attendances ADD COLUMN check_out_tz TIMESTAMPTZ;

-- 2. Migrar los datos combinando DATE + TIME
-- Usamos CASE para manejar campos nulos y evitar errores de conversión
UPDATE public.attendances SET 
    check_in_tz = (date + check_in)::TIMESTAMP WITH TIME ZONE,
    lunch_start_tz = CASE WHEN lunch_start IS NOT NULL THEN (date + lunch_start)::TIMESTAMP WITH TIME ZONE ELSE NULL END,
    lunch_end_tz = CASE WHEN lunch_end IS NOT NULL THEN (date + lunch_end)::TIMESTAMP WITH TIME ZONE ELSE NULL END,
    check_out_tz = CASE WHEN check_out IS NOT NULL THEN (date + check_out)::TIMESTAMP WITH TIME ZONE ELSE NULL END;

-- 3. Eliminar columnas antiguas y renombrar las nuevas
ALTER TABLE public.attendances DROP COLUMN check_in;
ALTER TABLE public.attendances DROP COLUMN lunch_start;
ALTER TABLE public.attendances DROP COLUMN lunch_end;
ALTER TABLE public.attendances DROP COLUMN check_out;
ALTER TABLE public.attendances DROP COLUMN date;

ALTER TABLE public.attendances RENAME COLUMN check_in_tz TO check_in;
ALTER TABLE public.attendances RENAME COLUMN lunch_start_tz TO lunch_start;
ALTER TABLE public.attendances RENAME COLUMN lunch_end_tz TO lunch_end;
ALTER TABLE public.attendances RENAME COLUMN check_out_tz TO check_out;
