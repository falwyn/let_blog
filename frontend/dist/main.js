"use strict";
console.log("TypeScript is compiled and running!");
const appRoot = document.getElementById("appRoot");
appRoot?.addEventListener('click', handlePostAction);
const createPostForm = document.getElementById("create-post-form");
createPostForm.addEventListener('submit', handleCreatePost);
const postTitle = document.getElementById("post-title");
const postContent = document.getElementById("post-content");
// Handler function when form is submitted
async function handleCreatePost(event) {
    // Stop browser from doing default action of reloading page when a form is submitted
    event.preventDefault();
    const titleValue = postTitle.value;
    const contentValue = postContent.value;
    await fetch('http://localhost:8080/posts/', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        // Convert the JavaScript object to a JSON string
        body: JSON.stringify({
            title: titleValue,
            content: contentValue,
        }),
    });
    postTitle.value = '';
    postContent.value = '';
    fetchAndRenderPosts();
}
async function handlePostAction(event) {
    const target = event.target;
    if (target.classList.contains('delete-btn')) {
        const postID = target.dataset.id;
        await fetch(`http://localhost:8080/posts/${postID}`, {
            method: 'DELETE',
        });
        fetchAndRenderPosts();
    }
    else if (target.classList.contains('edit-btn')) {
        const postElement = target.closest('div'); // find the main container for the post
        if (!postElement)
            return;
        const postTitle = postElement.querySelector('h2')?.innerText;
        const postContent = postElement.querySelector('p')?.innerText;
        const postID = target.dataset.id;
        // Replace post's content with editable form
        postElement.innerHTML = `
			<input type="text" class="edit-title" value="${postTitle}">
			<textarea class="edit-content">${postContent}</textarea>
			<button class="save-btn" data-id=${postID}>Save</button>
			<button class="cancel-btn">Cancel</button>
		`;
    }
    else if (target.classList.contains('save-btn')) {
        const postParentDiv = target.closest('div');
        if (!postParentDiv)
            return;
        const postID = target.dataset.id;
        const postElement = target.closest('div');
        const editedTitleInput = postElement?.querySelector('.edit-title');
        const editedContentTextarea = postElement?.querySelector('.edit-content');
        const editedTitle = editedTitleInput.value;
        const editedContent = editedContentTextarea.value;
        await fetch(`http://localhost:8080/posts/${postID}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                title: editedTitle,
                content: editedContent,
            }),
        }),
            fetchAndRenderPosts();
    }
    else if (target.classList.contains('cancel-btn')) {
        fetchAndRenderPosts();
    }
}
// Fetch and render posts
async function fetchAndRenderPosts() {
    try {
        // Make a GET request
        const response = await fetch("http://localhost:8080/posts/");
        // Parse the JSON response to an array of Post objects
        const posts = await response.json();
        // Clear any previous content
        if (appRoot) {
            appRoot.innerHTML = '';
        }
        if (posts === null && appRoot !== null) {
            appRoot.innerHTML = "<center>posts is null</center>";
        }
        else if (posts.length === 0 && appRoot !== null) {
            appRoot.innerHTML = "<center>Not blog posts yet</center>";
        }
        else {
            posts.forEach(post => {
                const postElement = document.createElement('div');
                postElement.innerHTML = `
      <h2>${post.title}</h2>
      <p>${post.content}</p>
			<button class="edit-btn" data-id="${post.id}">Edit</button>
			<button class="delete-btn" data-id="${post.id}">Delete</button>
      `;
                appRoot?.appendChild(postElement);
            });
        }
    }
    catch (error) {
        console.error("Failed to fetch posts: ", error);
        if (appRoot) {
            appRoot.innerHTML = "<p>Error loading posts. Is Go backend running?</p>";
        }
    }
}
fetchAndRenderPosts();
