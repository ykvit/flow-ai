import { render, screen, fireEvent } from '@testing-library/react';
import App from '../App';

describe('App Component', () => {
  it('renders the main heading and initial counter value', () => {
    render(<App />);
    expect(screen.getByText(/Flow-AI Frontend/i)).toBeInTheDocument();
    expect(screen.getByText(/Counter: 0/i)).toBeInTheDocument();
  });

  it('increments the counter when the button is clicked', () => {
    render(<App />);
    const button = screen.getByRole('button', { name: /Increment Counter/i });
    
    fireEvent.click(button);
    expect(screen.getByText(/Counter: 1/i)).toBeInTheDocument();

    fireEvent.click(button);
    expect(screen.getByText(/Counter: 2/i)).toBeInTheDocument();
  });
});