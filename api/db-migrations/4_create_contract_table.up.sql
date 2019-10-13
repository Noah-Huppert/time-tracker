CREATE TABLE contract (
	  id SERIAL PRIMARY KEY,
	  client_id INTEGER
	  		  NOT NULL
	  		  REFERENCES client
			  ON DELETE CASCADE,
	  employee_id INTEGER
	  		    NOT NULL
	  		    REFERENCES employee
			    ON DELETE CASCADE,
	  name TEXT NOT NULL,
	  hourly_rate FLOAT NOT NULL
)
