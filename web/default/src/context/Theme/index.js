import React, { useEffect, useCallback } from 'react';
import { initialState, reducer } from './reducer';

export const ThemeContext = React.createContext({
  state: initialState,
  dispatch: () => null,
  toggleTheme: () => null,
  setTheme: () => null,
});

export const ThemeProvider = ({ children }) => {
  const [state, dispatch] = React.useReducer(reducer, initialState);

  // Function to detect system theme preference
  const getSystemPreference = useCallback(() => {
    if (typeof window !== 'undefined' && window.matchMedia) {
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }
    return 'light';
  }, []);

  // Function to apply theme to document
  const applyTheme = useCallback((theme) => {
    const root = document.documentElement;
    const body = document.body;

    if (theme === 'dark') {
      root.setAttribute('data-theme', 'dark');
      body.classList.add('dark-theme');
      body.classList.remove('light-theme');
    } else {
      root.setAttribute('data-theme', 'light');
      body.classList.add('light-theme');
      body.classList.remove('dark-theme');
    }
  }, []);

  // Function to determine effective theme
  const getEffectiveTheme = useCallback((themePreference, systemPreference) => {
    if (themePreference === 'system') {
      return systemPreference;
    }
    return themePreference;
  }, []);

  // Set theme function
  const setTheme = useCallback((newTheme) => {
    dispatch({ type: 'SET_THEME', payload: newTheme });
    localStorage.setItem('theme', newTheme);

    const effectiveTheme = getEffectiveTheme(newTheme, state.systemPreference);
    dispatch({ type: 'SET_EFFECTIVE_THEME', payload: effectiveTheme });
    applyTheme(effectiveTheme);
  }, [state.systemPreference, getEffectiveTheme, applyTheme]);

  // Toggle theme function
  const toggleTheme = useCallback(() => {
    const newTheme = state.effectiveTheme === 'light' ? 'dark' : 'light';
    setTheme(newTheme);
  }, [state.effectiveTheme, setTheme]);

  // Initialize theme on mount
  useEffect(() => {
    const savedTheme = localStorage.getItem('theme') || 'system';
    const systemPreference = getSystemPreference();

    dispatch({ type: 'SET_SYSTEM_PREFERENCE', payload: systemPreference });
    dispatch({ type: 'SET_THEME', payload: savedTheme });

    const effectiveTheme = getEffectiveTheme(savedTheme, systemPreference);
    dispatch({ type: 'SET_EFFECTIVE_THEME', payload: effectiveTheme });
    applyTheme(effectiveTheme);
  }, [getSystemPreference, getEffectiveTheme, applyTheme]);

  // Listen for system theme changes with both event listener and periodic check
  useEffect(() => {
    if (typeof window !== 'undefined' && window.matchMedia) {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');

      const handleChange = (e) => {
        const newSystemPreference = e.matches ? 'dark' : 'light';
        dispatch({ type: 'SET_SYSTEM_PREFERENCE', payload: newSystemPreference });

        // Only update effective theme if user preference is 'system'
        if (state.theme === 'system') {
          dispatch({ type: 'SET_EFFECTIVE_THEME', payload: newSystemPreference });
          applyTheme(newSystemPreference);
        }
      };

      // Add event listener for immediate changes
      mediaQuery.addEventListener('change', handleChange);

      // Add periodic check every second for system theme changes
      // This ensures we catch changes even if the event listener fails
      const checkSystemTheme = () => {
        const currentSystemPreference = getSystemPreference();
        if (currentSystemPreference !== state.systemPreference) {
          dispatch({ type: 'SET_SYSTEM_PREFERENCE', payload: currentSystemPreference });

          // Only update effective theme if user preference is 'system'
          if (state.theme === 'system') {
            dispatch({ type: 'SET_EFFECTIVE_THEME', payload: currentSystemPreference });
            applyTheme(currentSystemPreference);
          }
        }
      };

      const intervalId = setInterval(checkSystemTheme, 1000); // Check every second

      return () => {
        mediaQuery.removeEventListener('change', handleChange);
        clearInterval(intervalId);
      };
    }
  }, [state.theme, state.systemPreference, applyTheme, getSystemPreference]);

  const value = {
    state,
    dispatch,
    toggleTheme,
    setTheme,
  };

  return (
    <ThemeContext.Provider value={value}>
      {children}
    </ThemeContext.Provider>
  );
};
