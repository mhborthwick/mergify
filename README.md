# Mergify

## Intro

A CLI (command-line interface) to merge your Spotify playlists.

## Example Playlists

| Name                           | User              | Description                                     | Link                                                                                            |
| ------------------------------ | ----------------- | ----------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| Mike's Monthly Playlists (All) | Mike H. Borthwick | A compilation of Mike's 100+ Monthly Playlists. | [Open on Spotify](https://open.spotify.com/playlist/4ID9qCTCaBebwOkMFgFd1o?si=e7e04f8a7a4344ba) |

## Requirements

- Docker (required for running the auth proxy server)

## Install

```sh
export ARCH="darwin-arm64" # macOS (Apple Silicon)

# export ARCH="darwin-amd64" # macOS (Intel)

# export ARCH="linux-amd64" # Linux (x86_64)

curl -o mergify -L "https://github.com/mhborthwick/mergify/releases/latest/download/mergify-${ARCH}"

chmod +x mergify

sudo mv mergify /usr/local/bin/
```

## Usage

```sh
Usage: mergify <command> [flags]

Flags:
  -h, --help    Show context-sensitive help.

Commands:
  create [flags]
    Combines the tracks from the playlists in your CLI config into a new playlist

Run "mergify <command> --help" for more information on a command.
```

## Setup

### Step 1: Authenticate with Spotify

#### 1.1 Create a Spotify App

You will need to create a Spotify app to obtain your client ID and secret for authentication.

- Visit the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)

- Create a new Spotify app

- In the app settings, add http://localhost:3000/callback to the `Redirect URIs` field

#### 1.2 Start the Auth Proxy Server

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

Now, start the auth proxy server using Docker Compose:

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

### Step 2: Set Up Your Mergify Config File

#### 2.1 Create Your Config File

Add a `~/.mergify/config.json` file:

```sh
touch ~/.mergify/config.json
```

#### 2.2 Define Your Playlists

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
