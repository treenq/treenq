CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    email varchar(85) NOT NULL,
    displayName varchar(255) NOT NULL UNIQUE,
    githubId integer NOT NULL,

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS installedRepos (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    githubId integer NOT NULL,
    fullName varchar(255) NOT NULL,
    private boolean NOT NULL,
    installationId integer NOT NULL,
    userId uuid REFERENCES users(id) NOT NULL,
    status varchar(25) NOT NULL,
    branch varchar(255) NOT NULL,

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS deployments (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    appId uuid NOT NULL,
    app jsonb NOT NULL,
    tag varchar(255) NOT NULL,
    sha varchar(255) NOT NULL,
    "user" uuid NOT NULL,

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

