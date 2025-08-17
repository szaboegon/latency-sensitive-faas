CREATE TABLE IF NOT EXISTS function_apps (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    runtime TEXT NOT NULL,
    components TEXT,
    files TEXT,
    source_path TEXT
);

CREATE TABLE IF NOT EXISTS function_compositions (
    id TEXT PRIMARY KEY,
    function_app_id TEXT NOT NULL,
    image TEXT,
    timestamp TEXT,
    files TEXT,           
    components TEXT, 
    status TEXT DEFAULT 'pending',     
    FOREIGN KEY (function_app_id) REFERENCES function_apps(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS deployments (
    id TEXT PRIMARY KEY,
    function_composition_id TEXT NOT NULL,
    node TEXT NOT NULL,
    namespace TEXT NOT NULL,
    routing_table TEXT NOT NULL,
    FOREIGN KEY (function_composition_id) REFERENCES function_compositions(id) ON DELETE CASCADE
);
