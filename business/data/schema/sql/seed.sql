INSERT INTO tasks (task_id, date_created, version, input_url, output_url, hooks, exec_image, timeout) VALUES
	('5cf37266-3473-4006-984f-9325122678b7', '2019-03-24 00:00:00', 'Test', 's3://june-test-bucket-jnk/dummy_cycling.mp4', 's3://processed-video/', 'mp4', 'jnkroeker/mp4_processor:0.1.4', '90'),
	('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', '2019-03-25 00:00:00', 'Test', 's3://june-test-bucket-jnk/dummy_skiing.mp4', 's3://processed-video/', 'mp4', 'jnkroeker/mp4_processor:0.1.4', '90')
	ON CONFLICT DO NOTHING;