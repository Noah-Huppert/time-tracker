CREATE TABLE project (
	  id SERIAL PRIMARY KEY,
	  client_id INTEGER
	  		  NOT NULL
	  		  REFERENCES client
			  ON DELETE CASCADE,
	  name TEXT NOT NULL,
	  description TEXT
)
