# rusodb

to create table: 
		CREATE TABLE dogs (name varchar, breed varchar);
    - only supports varchar as a type (this is temporary)

to insert: 
		INSERT INTO table_name (column1, column2) VALUES ("value1", "value2");
		or
		INSERT INTO table_name VALUES ("value1", "value2");

to create index on a column/s:
		CREATE INDEX ON wishlist (name, price);

to select: 
		SELECT * FROM dogs WHERE breed = "cane corso";
    - must use "*", for now, the full row will be returned (this is temporary, will be able to query for specific columns or the entire row soon)
