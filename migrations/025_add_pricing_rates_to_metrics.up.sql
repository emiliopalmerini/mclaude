ALTER TABLE session_metrics ADD COLUMN input_rate REAL;
ALTER TABLE session_metrics ADD COLUMN output_rate REAL;
ALTER TABLE session_metrics ADD COLUMN cache_read_rate REAL;
ALTER TABLE session_metrics ADD COLUMN cache_write_rate REAL;
