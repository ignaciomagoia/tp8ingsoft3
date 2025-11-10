const API_URL = process.env.REACT_APP_API_URL || "http://localhost:8080";

async function handleResponse(response) {
  const contentType = response.headers.get("Content-Type") || "";
  const isJSON = contentType.includes("application/json");
  const payload = isJSON ? await response.json().catch(() => ({})) : {};

  if (!response.ok) {
    const message = payload.error || "Error inesperado en el servidor";
    throw new Error(message);
  }

  return payload;
}

export async function registerUser({ email, password }) {
  const response = await fetch(`${API_URL}/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  return handleResponse(response);
}

export async function loginUser({ email, password }) {
  const response = await fetch(`${API_URL}/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  return handleResponse(response);
}

export async function getTodos(email) {
  const url = new URL(`${API_URL}/todos`);
  if (email) {
    url.searchParams.append("email", email);
  }
  const response = await fetch(url.toString());
  return handleResponse(response);
}

export async function createTodo({ email, title }) {
  const response = await fetch(`${API_URL}/todos`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, title }),
  });
  return handleResponse(response);
}

export async function updateTodo(id, data) {
  const response = await fetch(`${API_URL}/todos/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  return handleResponse(response);
}

export async function deleteTodo(id) {
  const response = await fetch(`${API_URL}/todos/${id}`, {
    method: "DELETE",
  });
  return handleResponse(response);
}
