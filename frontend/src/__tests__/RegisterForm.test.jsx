import React from "react";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import RegisterForm from "../components/RegisterForm";

describe("RegisterForm", () => {
  it("muestra errores cuando los campos están vacíos", async () => {
    render(<RegisterForm onRegister={jest.fn()} onLogin={jest.fn()} />);

    await act(async () => {
      await userEvent.click(screen.getByRole("button", { name: /registrar/i }));
    });

    expect(screen.getByText(/el email es requerido/i)).toBeInTheDocument();
    expect(screen.getByText(/la contraseña es requerida/i)).toBeInTheDocument();
  });

  it("llama a onRegister con los valores normalizados", async () => {
    const handleRegister = jest.fn();
    render(<RegisterForm onRegister={handleRegister} onLogin={jest.fn()} />);

    await act(async () => {
      await userEvent.type(screen.getByLabelText(/email/i), "  USER@Test.com ");
      await userEvent.type(screen.getByLabelText(/contraseña/i), "secret");
      await userEvent.click(screen.getByRole("button", { name: /registrar/i }));
    });

    expect(handleRegister).toHaveBeenCalledWith({
      email: "user@test.com",
      password: "secret",
    });
  });

  it("muestra un error cuando el email no es v\u00e1lido", async () => {
    render(<RegisterForm onRegister={jest.fn()} onLogin={jest.fn()} />);

    await act(async () => {
      await userEvent.type(screen.getByLabelText(/email/i), "usuario-sin-formato");
      await userEvent.type(screen.getByLabelText(/contrase\u00f1a/i), "pass");
      await userEvent.click(screen.getByRole("button", { name: /registrar/i }));
    });

    expect(
      screen.getByText(/el email no tiene un formato v\u00e1lido/i)
    ).toBeInTheDocument();
  });

  it("ejecuta onLogin y reinicia la contrase\u00f1a", async () => {
    const handleLogin = jest.fn();
    render(<RegisterForm onRegister={jest.fn()} onLogin={handleLogin} />);

    const emailInput = screen.getByLabelText(/email/i);
    const passwordInput = screen.getByLabelText(/contrase\u00f1a/i);

    await act(async () => {
      await userEvent.type(emailInput, "demo@example.com");
      await userEvent.type(passwordInput, "123456");
      await userEvent.click(screen.getByRole("button", { name: /iniciar sesi\u00f3n/i }));
    });

    expect(handleLogin).toHaveBeenCalledWith({
      email: "demo@example.com",
      password: "123456",
    });

    expect(passwordInput).toHaveValue("");
  });
});
