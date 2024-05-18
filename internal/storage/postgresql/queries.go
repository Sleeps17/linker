package postgresql

var (
	createUsersTableQuery = `CREATE TABLE IF NOT EXISTS "users" (
    	"id" SERIAL PRIMARY KEY, 
    	"username" TEXT UNIQUE NOT NULL
	);`
	selectUserQuery = `SELECT id FROM users WHERE username = $1;`
	insertUserQuery = `INSERT INTO users (username) VALUES ($1) RETURNING id;`
	deleteUserQuery = `DELETE FROM users WHERE username = $1;`

	createTopicsTableQuery = `CREATE TABLE IF NOT EXISTS "topics" (
    	"id" SERIAL PRIMARY KEY,
    	"user_id" INT NOT NULL,
    	"topic" TEXT NOT NULL,
    	FOREIGN KEY (user_id) REFERENCES users(id),
    	UNIQUE (user_id, topic)
	);`
	insertTopicQuery        = `INSERT INTO topics (user_id, topic) VALUES ($1, $2) RETURNING id;`
	deleteLinksByTopicQuery = `DELETE FROM links WHERE user_id = $1 AND topic_id = $2;`
	deleteTopicQuery        = `DELETE FROM topics WHERE user_id = $1 AND topic = $2 RETURNING id;`
	selectTopicQuery        = `SELECT id FROM topics WHERE user_id = $1 AND topic = $2;`
	listTopicsQuery         = `SELECT topic FROM topics WHERE user_id = $1;`

	createLinksTableQuery = `CREATE TABLE IF NOT EXISTS "links" (
    	"id" SERIAL PRIMARY KEY,
    	"user_id" INT NOT NULL,
    	"topic_id" INT NOT NULL,
    	"link" TEXT NOT NULL,
    	"alias" TEXT NOT NULL,
    	FOREIGN KEY (user_id) REFERENCES users(id),
    	FOREIGN KEY (topic_id) REFERENCES topics(id),
    	UNIQUE (user_id, topic_id, alias)
	);`
	insertLinkQuery = `INSERT INTO links (user_id, topic_id, link, alias) VALUES ($1, $2, $3, $4);`
	selectLinkQuery = `SELECT link FROm links WHERE user_id = $1 AND topic_id = $2 AND alias = $3;`
	listLinksQuery  = `SELECT link, alias FROM links WHERE user_id = $1 AND topic_id = $2;`
	deleteLinkQuery = `DELETE FROM links WHERE user_id = $1 AND topic_id = $2 AND alias = $3`
)
