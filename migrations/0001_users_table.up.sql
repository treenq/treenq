CREATE TABLE IF NOT EXISTS workspaces (
    id CHAR(20) PRIMARY KEY NOT NULL,
    name varchar(255) NOT NULL,
    githubOrgName varchar(255) NOT NULL,

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id CHAR(20) PRIMARY KEY  NOT NULL,
    email varchar(85) NOT NULL,
    displayName varchar(255) NOT NULL UNIQUE,

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS workspaceUsers (
    workspaceId CHAR(20) REFERENCES workspaces(id) NOT NULL,
    userId CHAR(20) REFERENCES users(id) NOT NULL,
    role varchar(20) NOT NULL DEFAULT 'member',

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY (workspaceId, userId)
);

CREATE TABLE IF NOT EXISTS installations (
    id CHAR(20) PRIMARY KEY NOT NULL,
    workspaceId CHAR(20) REFERENCES workspaces(id) NOT NULL,
    githubId integer NOT NULL UNIQUE,

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS installedRepos (
    id CHAR(20) PRIMARY KEY NOT NULL,
    githubId integer NOT NULL UNIQUE,
    fullName varchar(255) NOT NULL,
    private boolean NOT NULL,
    installationId integer NOT NULL,
    workspaceId CHAR(20) REFERENCES workspaces(id) NOT NULL,
    status varchar(25) NOT NULL,
    branch varchar(100) NOT NULL,

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS deployments (
    id CHAR(20) PRIMARY KEY NOT NULL,
    fromDeploymentId CHAR(20) NOT NULL,
    repoId CHAR(20) NOT NULL,

    space jsonb NOT NULL,
    sha char(40) NOT NULL,
    branch  varchar(100) NOT NULL,
    commitMessage text NOT NULL,
    buildTag varchar(80),
    userDisplayName varchar(255) NOT NULL,
    status varchar(24) NOT NULL,

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS secrets (
    repoId CHAR(20) NOT NULL REFERENCES installedRepos(id),
    key varchar(64) NOT NULL,
    workspaceId CHAR(20) REFERENCES workspaces(id) NOT NULL,
    createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (repoId, key)
);

CREATE TABLE IF NOT EXISTS spaces (
    repoId CHAR(20) PRIMARY KEY NOT NULL REFERENCES installedRepos(id),
    space jsonb NOT NULL,
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
