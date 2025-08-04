CREATE TABLE IF NOT EXISTS function_apps (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS function_compositions (
    id TEXT PRIMARY KEY,
    function_app_id TEXT NOT NULL,
    node TEXT,
    namespace TEXT NOT NULL,
    source_path TEXT,
    runtime TEXT NOT NULL,
    image TEXT,
    timestamp TEXT,
    files TEXT,           
    components TEXT,      
    FOREIGN KEY (function_app_id) REFERENCES function_apps(id) ON DELETE CASCADE
);
