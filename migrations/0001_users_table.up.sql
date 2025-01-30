CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    email varchar(85) NOT NULL,
    displayName varchar(255),

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
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

CREATE TABLE IF NOT EXISTS repos (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    fullName varchar(255) NOT NULL,
    email varchar(85) NOT NULL,
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS authStates (
    email varchar(85) NOT NULL UNIQUE,
    state varchar(255) NOT NULL,
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS githubTokens (
    email varchar(85) NOT NULL PRIMARY KEY,
    accessToken varchar(255) NOT NULL,
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS githubRepos (
    email varchar(85) NOT NULL,
    repoId varchar(255) NOT NULL,
    fullName varchar(255) NOT NULL,
    defaultBranch varchar(255) NOT NULL,
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
