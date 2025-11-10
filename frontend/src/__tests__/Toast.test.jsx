import React from "react";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import Toast from "../components/Toast";

describe("Toast", () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  it("no renderiza nada cuando no hay mensaje", () => {
    render(<Toast message="" type="info" onClose={jest.fn()} />);
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("cierra autom\u00e1ticamente y manualmente", async () => {
    const handleClose = jest.fn();
    render(<Toast message="Guardado" type="success" onClose={handleClose} />);

    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent(/guardado/i);

    await userEvent.click(screen.getByRole("button", { name: /cerrar notificaci\u00f3n/i }));
    expect(handleClose).toHaveBeenCalledTimes(1);

    await act(async () => {
      jest.advanceTimersByTime(3500);
    });

    expect(handleClose).toHaveBeenCalledTimes(2);
  });
});
