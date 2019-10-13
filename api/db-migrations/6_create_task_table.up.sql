CREATE TABLE task (
	  id SERIAL PRIMARY KEY,
	  time_entry_id INTEGER
	  			 NOT NULL
	  			 REFERENCES time_entry
				 ON DELETE CASCADE,
	  title TEXT NOT NULL,
	  description TEXT
)
