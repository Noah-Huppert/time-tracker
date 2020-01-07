CREATE TABLE task (
	  id SERIAL PRIMARY KEY,
	  title TEXT NOT NULL,
	  description TEXT NOT NULL,
	  project_id INTEGER
	  		   NOT NULL
	  		   REFERENCES project
			   ON DELETE CASCADE,
       contract_id INTEGER
	  		    NOT NULL
	  		    REFERENCES contract
			    ON DELETE CASCADE,
	  start_on TIMESTAMP WITH TIME ZONE NOT NULL,
	  end_on TIMESTAMP WITH TIME ZONE,
	  duration INTERVAL GENERATED ALWAYS AS (end_on - start_on) STORED
)
