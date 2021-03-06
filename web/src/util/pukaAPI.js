const JSON_API_CONTENT_TYPE = 'application/vnd.api+json';

function mapResponse(response) {
  const bookmarks = {};
  try {
    if (Array.isArray(response.data)) {
      for (const item of response.data) {
        item.attributes.id = item.id;
        bookmarks[item.id] = item.attributes;
      }
    } else {
      const item = response.data;
      item.attributes.id = item.id;
      bookmarks[item.id] = item.attributes;
    }
  } catch (e) {
    return Promise.reject(`Error in pukaAPI.mapResponse: ${e.message}`);
  }
  return Promise.resolve({
    entities: { bookmarks },
    nextPage: response.links ? response.links.next : null,
  });
}

export const getBookmarks = (endpoint) =>
  fetch(endpoint, {
    redirect: 'follow',
    headers: {
      Accept: JSON_API_CONTENT_TYPE,
    },
  }).then(response => {
    if (response.status >= 200 && response.status < 300) {
      return Promise.resolve(response.json());
    }
    return Promise.reject(`Error in pukaAPI.getBookmarks response: ${response.statusText}`);
  })
  .then(mapResponse);

export const saveBookmark = (endpoint, bookmark) => {
  const body = {
    data: {
      type: 'bookmarks',
      id: bookmark.id,
      attributes: bookmark,
    },
  };
  if (typeof body.data.attributes.tags === 'string') {
    body.data.attributes.tags = bookmark.tags.split(/\s*,\s*/);
  }
  const token = localStorage.getItem('puka_auth_token');
  return fetch(`${endpoint}?token=${token}`, {
    method: bookmark.id ? 'PATCH' : 'POST',
    body: JSON.stringify(body),
    redirect: 'follow',
    headers: {
      Accept: JSON_API_CONTENT_TYPE,
      'Content-Type': JSON_API_CONTENT_TYPE,
    },
  }).then(response => {
    if (response.status >= 200 && response.status < 300) {
      return Promise.resolve(response.json());
    }
    return Promise.reject(`Error in pukaAPI.saveBookmark response: ${response.statusText}`);
  })
  .then(mapResponse);
};

export const updateBookmark = (endpoint, bookmark) =>
  saveBookmark(`${endpoint}/${bookmark.id}`, bookmark);
