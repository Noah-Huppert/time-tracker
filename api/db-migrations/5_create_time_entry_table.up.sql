CREATE TABLE time_entry (
	  id SERIAL PRIMARY KEY,
	  project_id INTEGER
	  		   NOT NULL
	  		   REFERENCES project
			   ON DELETE CASCADE,
       contract_id INTEGER
	  		    NOT NULL
	  		    REFERENCES contract
			    ON DELETE CASCADE,
	  hours FLOAT NOT NULL
)
