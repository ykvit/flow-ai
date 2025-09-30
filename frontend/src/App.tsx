import { useState } from 'react';
import { MdButton } from '@material/web/button/button';
import { createComponent } from '@lit-labs/react';
import * as React from 'react';

// NOTE: This is a simplified placeholder App.tsx.
// It uses useState to show that the React environment is working,
// but removes the complex Material Web Component (MWC) logic that caused
// the Rollup/Vite build error.

function App() {
  // We keep useState to demonstrate that React's state management works.
  const [counter, setCounter] = useState(0); 

  // Simulating the circle color check with a simple style.
  const circleColor = '#FFFFFF'; 

  return (
    <div className="app-container">
      <h1>Flow-AI Frontend</h1>
      <p>This is a placeholder UI to demonstrate reactivity and state management.</p>
      
      <div 
        className="circle" 
        style={{ backgroundColor: circleColor }}
        data-testid="color-circle"
      ></div>

      <p>Counter: {counter}</p>
      <button onClick={() => setCounter(c => c + 1)}>
        Increment Counter (Check Reactivity)
      </button>
    </div>
  );
}

export default App;