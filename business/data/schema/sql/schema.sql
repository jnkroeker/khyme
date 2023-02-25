-- Version:1.1
-- Description: Create table tasks
CREATE TABLE tasks (
    task_id         UUID,        
	date_created    TIMESTAMP,  
	version         TEXT,  
	input_url       TEXT,     
	output_url      TEXT,      
	hooks           TEXT,       
	exec_image      TEXT,       
	timeout         INT,

    PRIMARY KEY (task_id)
);