import { useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';

import { ThemeProvider } from '@mui/material/styles';
import { CssBaseline, StyledEngineProvider } from '@mui/material';
import { SET_THEME } from 'store/actions';
// routing
import Routes from 'routes';

// defaultTheme
import themes from 'themes';

// project imports
import NavigationScroll from 'layout/NavigationScroll';

// auth
import UserProvider from 'contexts/UserContext';
import StatusProvider from 'contexts/StatusContext';
import { SnackbarProvider } from 'notistack';

// ==============================|| APP ||============================== //

const App = () => {
  const dispatch = useDispatch();
  const customization = useSelector((state) => state.customization);

  useEffect(() => {
    const storedTheme = localStorage.getItem('theme') || 'system';

    let effectiveTheme = storedTheme;
    if (storedTheme === 'system') {
      // Detect system preference
      if (typeof window !== 'undefined' && window.matchMedia) {
        effectiveTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
      } else {
        effectiveTheme = 'light';
      }
    }

    dispatch({ type: SET_THEME, theme: effectiveTheme });

    // Listen for system theme changes when in system mode
    if (storedTheme === 'system' && typeof window !== 'undefined' && window.matchMedia) {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');

      const handleChange = (e) => {
        if (localStorage.getItem('theme') === 'system') {
          const newTheme = e.matches ? 'dark' : 'light';
          dispatch({ type: SET_THEME, theme: newTheme });
        }
      };

      mediaQuery.addEventListener('change', handleChange);
      return () => mediaQuery.removeEventListener('change', handleChange);
    }
  }, [dispatch]);

  return (
    <StyledEngineProvider injectFirst>
      <ThemeProvider theme={themes(customization)}>
        <CssBaseline />
        <NavigationScroll>
          <SnackbarProvider autoHideDuration={5000} maxSnack={3} anchorOrigin={{ vertical: 'top', horizontal: 'right' }}>
            <UserProvider>
              <StatusProvider>
                <Routes />
              </StatusProvider>
            </UserProvider>
          </SnackbarProvider>
        </NavigationScroll>
      </ThemeProvider>
    </StyledEngineProvider>
  );
};

export default App;
