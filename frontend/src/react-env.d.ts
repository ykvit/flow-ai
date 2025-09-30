// This file extends TypeScript's built-in JSX types to include
// Material Web Components. Without this, TypeScript would not recognize
// tags like <md-button> and would throw a compilation error.

// We import React to access its type definitions.
import * as React from 'react';

// We declare a global module to augment existing types.
declare global {
  // We enter the JSX namespace to modify its elements.
  namespace JSX {
    // IntrinsicElements is an interface that lists all known HTML tags.
    // We are adding our custom elements to this list.
    interface IntrinsicElements {
      // This tells TypeScript that 'md-button' is a valid JSX tag.
      // It should be treated like a standard HTML element.
      'md-button': React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement>;
      
      // As you add more Material components (e.g., <md-icon>, <md-text-field>),
      // you will add them here as well.
    }
  }
}