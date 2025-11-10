const TODOS_API = '**/todos*';
const LOGIN_API = '**/login';
const USER = { email: 'cypress@example.com', password: 'superSecret123' };

const loginAndLoad = (initialTodos = []) => {
  cy.intercept('POST', LOGIN_API, {
    statusCode: 200,
    body: { message: 'Inicio de sesión exitoso' },
  }).as('login');

  cy.intercept('GET', TODOS_API, {
    statusCode: 200,
    body: { todos: initialTodos },
  }).as('loadTodos');

  cy.visit('/');

  cy.get('input[aria-label="Email"]').clear().type(USER.email);
  cy.get('input[aria-label="Contraseña"]').clear().type(USER.password);
  cy.contains('button', /Iniciar sesión/i).click();

  cy.wait('@login');
  cy.wait('@loadTodos');
};

describe('Todos – flujos principales', () => {
  it('Crear: agrega una tarea y se ve en la lista', () => {
    loginAndLoad([]);

    cy.intercept('POST', '**/todos', (req) => {
      expect(req.body.title).to.contain('Comprar pan');
      req.reply({
        statusCode: 201,
        body: { todo: { id: '1', title: req.body.title, completed: false } },
      });
    }).as('createTodo');

    cy.get('[data-cy=todo-input]').type('Comprar pan');
    cy.get('[data-cy=add-btn]').click();

    cy.wait('@createTodo');
    cy.contains('[data-cy=todo-item]', 'Comprar pan').should('be.visible');
  });

  it('Editar: modifica el título de una tarea existente', () => {
    loginAndLoad([{ id: '10', title: 'Original', completed: false }]);

    cy.intercept('PUT', '**/todos/10', (req) => {
      expect(req.body.title).to.eq('Editada');
      req.reply({
        statusCode: 200,
        body: { todo: { id: '10', title: 'Editada', completed: false } },
      });
    }).as('updateTodo');

    cy.get('[data-cy=edit-10]').click();
    cy.get('[data-cy="edit-input-10"]').clear().type('Editada');
    cy.get('[data-cy=save-10]').click();

    cy.wait('@updateTodo');
    cy.contains('[data-cy=todo-item]', 'Editada').should('be.visible');
  });

  it('Error: muestra feedback cuando la API falla', () => {
    loginAndLoad([]);

    cy.intercept('POST', '**/todos', { statusCode: 500, body: { error: 'Boom' } }).as('createFail');

    cy.get('[data-cy=todo-input]').type('Falla controlada');
    cy.get('[data-cy=add-btn]').click();
    cy.wait('@createFail');

    cy.contains(/boom|error|falló|failed|intente|try again/i).should('be.visible');
  });
});
