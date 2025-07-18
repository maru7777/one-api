import { useDispatch, useSelector } from 'react-redux';
import { SET_THEME } from 'store/actions';
import { useTheme } from '@mui/material/styles';
import { Avatar, Box, ButtonBase } from '@mui/material';
import { IconSun, IconMoon, IconDeviceDesktop } from '@tabler/icons-react';

export default function ThemeButton() {
  const dispatch = useDispatch();

  const defaultTheme = useSelector((state) => state.customization.theme);

  const theme = useTheme();

  return (
    <Box
      sx={{
        ml: 2,
        mr: 3,
        [theme.breakpoints.down('md')]: {
          mr: 2
        }
      }}
    >
      <ButtonBase sx={{ borderRadius: '12px' }}>
        <Avatar
          variant="rounded"
          sx={{
            ...theme.typography.commonAvatar,
            ...theme.typography.mediumAvatar,
            transition: 'all .2s ease-in-out',
            borderColor: theme.typography.menuChip.background,
            backgroundColor: theme.typography.menuChip.background,
            '&[aria-controls="menu-list-grow"],&:hover': {
              background: theme.palette.secondary.dark,
              color: theme.palette.secondary.light
            }
          }}
          onClick={() => {
            const currentStoredTheme = localStorage.getItem('theme') || 'system';
            let newTheme;
            let effectiveTheme;

            // Cycle through: light -> dark -> system -> light
            if (currentStoredTheme === 'light') {
              newTheme = 'dark';
              effectiveTheme = 'dark';
            } else if (currentStoredTheme === 'dark') {
              newTheme = 'system';
              // Get system preference for effective theme
              if (typeof window !== 'undefined' && window.matchMedia) {
                effectiveTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
              } else {
                effectiveTheme = 'light';
              }
            } else {
              newTheme = 'light';
              effectiveTheme = 'light';
            }

            localStorage.setItem('theme', newTheme);
            dispatch({ type: SET_THEME, theme: effectiveTheme });
          }}
          color="inherit"
        >
          {(() => {
            const storedTheme = localStorage.getItem('theme') || 'system';
            if (storedTheme === 'system') {
              return <IconDeviceDesktop stroke={1.5} size="1.3rem" />;
            }
            return defaultTheme === 'light' ? <IconSun stroke={1.5} size="1.3rem" /> : <IconMoon stroke={1.5} size="1.3rem" />;
          })()}
        </Avatar>
      </ButtonBase>
    </Box>
  );
}
