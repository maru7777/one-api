export const initialState = {
  theme: 'light', // 'light' | 'dark' | 'system'
  effectiveTheme: 'light', // actual theme being applied
  systemPreference: 'light'
};

export const reducer = (state, action) => {
  switch (action.type) {
    case 'SET_THEME':
      return {
        ...state,
        theme: action.payload
      };
    case 'SET_EFFECTIVE_THEME':
      return {
        ...state,
        effectiveTheme: action.payload
      };
    case 'SET_SYSTEM_PREFERENCE':
      return {
        ...state,
        systemPreference: action.payload
      };
    default:
      return state;
  }
};
