# Mergify

## Intro

Mergify is a CLI (command-line interface) that allows you to merge your Spotify playlists.

## Example Playlists

| Name                           | User              | Description                                     | Link                                                                                            |
| ------------------------------ | ----------------- | ----------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| Mike's Monthly Playlists (All) | Mike H. Borthwick | A compilation of Mike's 100+ Monthly Playlists. | [Open on Spotify](https://open.spotify.com/playlist/4ID9qCTCaBebwOkMFgFd1o?si=e7e04f8a7a4344ba) |

## Requirements

- Git

- Docker

## Install

```sh
export ARCH="darwin-arm64"

# export ARCH="darwin-amd64"

# export ARCH="linux-amd64"

# export ARCH="windows-amd64"

curl -o mergify -L "https://github.com/mhborthwick/mergify/releases/latest/download/mergify-${ARCH}"

chmod +x mergify

sudo mv mergify /usr/local/bin/
```

## Setup

### Overview

Before using the Mergify CLI, ensure you've obtained an access token (required to access your Spotify account) and set up your Mergify config file using the steps below.

### Step 1: Obtain an Access Token

#### 1.1 Create a Spotify App

You will need to create a Spotify app to obtain your client ID and secret for authentication.

- Visit the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)

- Create a new Spotify app

- In the app settings, add http://localhost:3000/callback to the `Redirect URIs` field

#### 1.2 Start the Authentication Server

Clone this repo:

```sh
git clone git@github.com:mhborthwick/mergify.git
```

Go to the `auth` directory and create a `.env` file:

```sh
cd mergify/auth

touch .env
```

Add your client ID and secret to `.env`:

```sh
CLIENT_ID=<replace_with_your_client_id>
CLIENT_SECRET=<replace_with_your_client_secret>
```

Now, start the auth server using Docker Compose:

```sh
make auth-up
```

Alternatively, run the following commands:

```sh
cd mergify/auth

docker compose up
```

#### 1.3 Generate an Access Token

The auth server runs at `http://localhost:3000`.

1. Open `http://localhost:3000` in your browser.

1. Click `Login with Spotify`.

1. Log in to your Spotify account and grant access, if prompted.

1. You’ll be redirected to a page displaying your access token.

1. Click `Copy Access Token` to save it to your clipboard (You'll use this in **Step 2: Set Up Your Mergify Config File** below).

> Note: Your token expires after ~30 minutes. Revisit `http://localhost:3000` to regenerate it if needed.

### Step 2: Set Up Your Mergify Config File

#### 2.1 Create Your Config File

Add a `~/.mergify/config.json` file:

```sh
touch ~/.mergify/config.json
```

#### 2.2 Add Your Playlists

Define the playlists that you want to merge as an array:

```jsonc
{
  "playlists": [
    "Playlist 1",
    "Playlist 2",
    "Playlist 3",
    "Playlist 4",
    "Playlist 5"
    // ...
  ]
}
```

> Note: Ensure the names match exactly as they appear in your Spotify account.

#### 2.3 Add Your Access Token

Copy your access token and set it as the value of `"token"` in your `~/.mergify/config.json` file:

```jsonc
{
  "token": "<replace_with_your_token>",
  "playlists": [
    // ...
  ]
}
```

> Note: If your token ever expires, replace it with a fresh one.

## Usage

### Create

The `create` command combines the tracks from the playlists defined in your Mergify config into a new playlist.

```sh
mergify create
```
