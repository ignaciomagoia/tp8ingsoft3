import {
  registerUser,
  loginUser,
  getTodos,
  createTodo,
  updateTodo,
  deleteTodo,
} from "../services/api";

describe("api service", () => {
  const originalFetch = global.fetch;

  beforeEach(() => {
    global.fetch = jest.fn();
  });

  afterEach(() => {
    jest.resetAllMocks();
    global.fetch = originalFetch;
  });

  const mockResponse = ({ ok = true, json = () => Promise.resolve({}), headers }) => ({
    ok,
    json,
    headers: {
      get: () => (headers?.["Content-Type"] ?? "application/json"),
    },
  });

  it("env\u00eda las credenciales de registro al endpoint correspondiente", async () => {
    global.fetch.mockResolvedValue(
      mockResponse({ json: () => Promise.resolve({ message: "Registrado" }) })
    );

    const payload = { email: "user@example.com", password: "secret" };
    await registerUser(payload);

    expect(global.fetch).toHaveBeenCalledWith("http://localhost:8080/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
  });

  it("adjunta el email como query string al solicitar todos", async () => {
    global.fetch.mockResolvedValue(
      mockResponse({ json: () => Promise.resolve({ todos: [] }) })
    );

    await getTodos("demo@example.com");

    expect(global.fetch).toHaveBeenCalledWith(
      "http://localhost:8080/todos?email=demo%40example.com"
    );
  });

  it("propaga un error con el mensaje del backend cuando la respuesta no es OK", async () => {
    global.fetch.mockResolvedValue(
      mockResponse({
        ok: false,
        json: () => Promise.resolve({ error: "Credenciales inv\u00e1lidas" }),
      })
    );

    await expect(loginUser({ email: "demo@example.com", password: "bad" })).rejects.toThrow(
      "Credenciales inv\u00e1lidas"
    );
  });

  it("usa un mensaje por defecto si la respuesta err\u00f3nea no es JSON", async () => {
    global.fetch.mockResolvedValue(
      mockResponse({
        ok: false,
        headers: { "Content-Type": "text/plain" },
        json: () => Promise.reject(new Error("fallo parsing")),
      })
    );

    await expect(createTodo({ email: "demo@example.com", title: "Test" })).rejects.toThrow(
      "Error inesperado en el servidor"
    );
  });

  it("propaga correctamente las llamadas de actualizaci\u00f3n y eliminaci\u00f3n", async () => {
    global.fetch
      .mockResolvedValueOnce(
        mockResponse({ json: () => Promise.resolve({ todo: { id: "1", completed: true } }) })
      )
      .mockResolvedValueOnce(mockResponse({ json: () => Promise.resolve({ message: "ok" }) }));

    await updateTodo("1", { completed: true });
    await deleteTodo("1");

    expect(global.fetch).toHaveBeenNthCalledWith(1, "http://localhost:8080/todos/1", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ completed: true }),
    });
    expect(global.fetch).toHaveBeenNthCalledWith(2, "http://localhost:8080/todos/1", {
      method: "DELETE",
    });
  });
});
