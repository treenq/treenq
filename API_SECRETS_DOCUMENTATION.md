# Secrets API Documentation

## Overview

The Secrets API allows users to securely manage sensitive information (secrets) associated with their repositories. This includes creating, updating, listing, and retrieving secrets. Secret values are stored encrypted in a secure backend (Kubernetes), while their keys are registered in the application's database for reference.

## Authentication

All requests to the Secrets API must be authenticated. The API uses standard Bearer Token authentication. Include an `Authorization` header with your API requests:

`Authorization: Bearer <your_access_token>`

## Endpoints

---

### 1. Set Secret

Creates or updates a secret for a repository. If a secret with the given key already exists, its value will be updated. Otherwise, a new secret will be created.

*   **HTTP Method**: `POST`
*   **Path**: `/repositories/{repoID}/secrets`
*   **Description**: Stores a secret's key-value pair. The key is registered in the database, and the value is securely stored in Kubernetes.
*   **Path Parameters**:
    *   `repoID` (string, required): The unique identifier for the repository.
*   **Request Body**: `application/json`
    *   `secret_key` (string, required): The name of the secret key.
    *   `secret_value` (string, required): The value of the secret.
    ```json
    {
      "secret_key": "MY_API_KEY",
      "secret_value": "supersecretvalue123"
    }
    ```
*   **Response**:
    *   `201 Created`: Secret was successfully created or updated. The response body is empty.
*   **Possible Status Codes**:
    *   `201 Created`: Secret stored/updated successfully.
    *   `400 Bad Request`: The request was malformed (e.g., missing `secret_key` or `secret_value`, or they are empty).
    *   `401 Unauthorized`: Authentication token is missing or invalid.
    *   `404 Not Found`: The specified `repoID` does not exist. (Note: Depending on system design, this might also manifest as a general error if repository existence is not explicitly checked before secret operation).
    *   `500 Internal Server Error`: An unexpected error occurred on the server.

---

### 2. Get Secret Keys

Retrieves a list of all registered secret keys for a specific repository. This operation does not return the actual secret values.

*   **HTTP Method**: `GET`
*   **Path**: `/repositories/{repoID}/secrets`
*   **Description**: Lists the keys of all secrets associated with the given repository.
*   **Path Parameters**:
    *   `repoID` (string, required): The unique identifier for the repository.
*   **Request Body**: None
*   **Response**: `200 OK`
    *   Returns a JSON array of strings, where each string is a secret key.
    ```json
    [
      "MY_API_KEY",
      "DATABASE_URL_STAGING",
      "THIRD_PARTY_SERVICE_TOKEN"
    ]
    ```
    *   If no secrets are registered, an empty array `[]` is returned.
*   **Possible Status Codes**:
    *   `200 OK`: Successfully retrieved the list of secret keys.
    *   `401 Unauthorized`: Authentication token is missing or invalid.
    *   `404 Not Found`: The specified `repoID` does not exist. (Note: As above, repository validation might affect this).
    *   `500 Internal Server Error`: An unexpected error occurred on the server.

---

### 3. View Secret Value

Retrieves the actual value of a specific secret for a repository, identified by its key.

*   **HTTP Method**: `GET`
*   **Path**: `/repositories/{repoID}/secrets/{secretKey}`
*   **Description**: Fetches the decrypted value of a specific secret.
*   **Path Parameters**:
    *   `repoID` (string, required): The unique identifier for the repository.
    *   `secretKey` (string, required): The key of the secret whose value is to be retrieved.
*   **Request Body**: None
*   **Response**: `200 OK`
    *   Returns a JSON object containing the secret's value.
    ```json
    {
      "secret_value": "supersecretvalue123"
    }
    ```
*   **Possible Status Codes**:
    *   `200 OK`: Successfully retrieved the secret value.
    *   `401 Unauthorized`: Authentication token is missing or invalid.
    *   `404 Not Found`: The specified `repoID` does not exist, or the `secretKey` does not exist for the given repository.
    *   `500 Internal Server Error`: An unexpected error occurred on the server (e.g., failure to decrypt the secret or a problem with the Kubernetes backend).

---
