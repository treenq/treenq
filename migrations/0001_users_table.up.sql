CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    email varchar(85) NOT NULL,
    displayName varchar(255) NOT NULL UNIQUE,

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS installations (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    userId uuid REFERENCES users(id) NOT NULL,
    githubId integer NOT NULL,

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS installedRepos (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    githubId integer NOT NULL,
    fullName varchar(255) NOT NULL,
    private boolean NOT NULL,
    installationId integer NOT NULL,
    userId uuid REFERENCES users(id) NOT NULL,
    status varchar(25) NOT NULL,
    branch varchar(100) NOT NULL,

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS deployments (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    repoId uuid NOT NULL,

    space jsonb NOT NULL,
    sha char(64) NOT NULL,
    buildTag varchar(80),
    userDisplayName varchar(255) NOT NULL,

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

