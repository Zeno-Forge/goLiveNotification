# Go Live Notification Scheduler

This application allows users to schedule and send notifications to Discord via webhooks. It supports scheduling posts with customizable messages, embeds, images, and thumbnails.

## Features

- **Schedule Posts:** Schedule posts to be sent to Discord at a specified time.
- **Customizable Messages:** Customize the content, title, description, URL, color, thumbnail, image, and footer of the Discord embed.
- **Image and Thumbnail Uploads:** Upload images and thumbnails to be included in the Discord embed.
- **Template Editing:** Edit a default post template to be used for new posts.

## Tech Stack

- **Go:** Backend programming language.
- **Echo:** High-performance, extensible, minimalist Go web framework.
- **templ:** A simple, fast, and type-safe HTML templating engine for Go.
- **HTMX:** Allows access to AJAX, CSS Transitions, WebSockets, and Server Sent Events directly in HTML.
- **Tailwind CSS:** A utility-first CSS framework.

## Setup

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/Zeno-Forge/goLiveNotification.git
    cd goLiveNotification
    ```

2.  **Install Dependencies:**

    ```bash
    go mod download
    ```

3.  **Build**

    ```bash
    go build -o goLiveNotif
    ```

4.  **Run the application:**

    ```bash
    ./goLiveNotif
    ```

    Or with Docker

    ```bash
    docker build . -t golivenotif
    docker run -p 8080:8080 golivenotif
    ```

5.  **Environment Variables**
    - `GOLIVE_PORT`: The port the application will run on. Defaults to `:8080`.
      ```Dockerfile
      startLine: 16
      endLine: 16
      ```

The application will be accessible at `http://localhost:8080` (or the port specified in `GOLIVE_PORT`).

## Usage

1.  **Access the Application:** Open a web browser and navigate to `http://localhost:8080`.
2.  **Settings:** Click the settings icon (⚙️) in the header to access the settings slide-out menu, where you can configure the Discord Webhook URL.
3.  **Edit Template:** Click the settings icon in the top right, and then click the "Edit Template" button to define a default post template.
4.  **Create a New Post:** Click the "Create New Post" button to open the post creation modal. The post will push your post as an embedded message to Discord via webhook when the scheduled time is reached.
5.  **Edit a Post:** Click the "Edit" button on a post in the list to open the post editing modal.
6.  **Delete a Post:** Click the "Delete" button on a post in the list to delete it.

## Contributing

This is currently a personal project, incidents of bugs or feature requests are welcome.
