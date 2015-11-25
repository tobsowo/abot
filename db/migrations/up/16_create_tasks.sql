CREATE TABLE tasks (
	id SERIAL,
	resultid INTEGER,
	userid INTEGER NOT NULL,
	resulttable VARCHAR(255) NOT NULL,
	packagename VARCHAR(255) NOT NULL,
	createdat TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	completedat TIMESTAMP,
	PRIMARY KEY (id)
);
