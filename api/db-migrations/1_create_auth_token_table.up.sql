CREATE TABLE auth_token (
	  id SERIAL PRIMARY KEY,
	  employee_id INTEGER
	  		    NOT NULL
	  	  	    REFERENCES employee
			    ON DELETE CASCADE,
	  name TEXT NOT NULL,
	  hash TEXT NOT NULL
)
