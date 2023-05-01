CREATE TABLE vehicle_locations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    route_tag TEXT NOT NULL,
    dir_tag TEXT NOT NULL,
    vehicle_id INTEGER NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    speed INTEGER NOT NULL,
    age INTEGER NOT NULL,
    heading INTEGER NOT NULL,
    creation_timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);

