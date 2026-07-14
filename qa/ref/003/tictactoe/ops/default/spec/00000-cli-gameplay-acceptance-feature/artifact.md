Feature: CLI tic-tac-toe gameplay

  Scenario: A player wins a valid game
    Given a new tic-tac-toe game
    When X plays the top row
    Then the CLI reports that X won

