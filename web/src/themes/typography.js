/**
 * @param {JsonObject} theme theme customization object
 */

const primaryFont =
  '"Public Sans Variable",-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif,"Apple Color Emoji","Segoe UI Emoji","Segoe UI Symbol"';
const secondaryFont =
  '"Barlow",-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif,"Apple Color Emoji","Segoe UI Emoji","Segoe UI Symbol"';

export default function themeTypography(theme) {
  return {
    fontFamily: theme?.customization?.fontFamily || primaryFont,
    fontSecondaryFamily: secondaryFont,
    fontWeightLight: 300,
    fontWeightRegular: 400,
    fontWeightMedium: 500,
    fontWeightSemiBold: 600,
    fontWeightBold: 700,
    h1: {
      fontWeight: 800,
      fontSize: '2.125rem',
      lineHeight: 1.35,
      fontFamily: secondaryFont,
      color: theme.heading
    },
    h2: {
      fontWeight: 800,
      fontSize: '1.5rem',
      lineHeight: 1.35,
      fontFamily: secondaryFont,
      color: theme.heading
    },
    h3: {
      fontWeight: 700,
      fontSize: '1.25rem',
      lineHeight: 1.5,
      fontFamily: secondaryFont,
      color: theme.heading
    },
    h4: {
      fontWeight: 700,
      fontSize: '1.125rem',
      lineHeight: 1.5,
      color: theme.heading
    },
    h5: {
      fontWeight: 600,
      fontSize: '1rem',
      lineHeight: 1.5,
      color: theme.heading
    },
    h6: {
      fontWeight: 600,
      fontSize: '0.875rem',
      lineHeight: 28 / 18,
      color: theme.heading
    },
    subtitle1: {
      fontSize: '1rem',
      lineHeight: 1.5,
      fontWeight: 600,
      color: theme.textDark
    },
    subtitle2: {
      fontSize: '0.875rem',
      lineHeight: 22 / 14,
      fontWeight: 600,
      color: theme.darkTextSecondary
    },
    body1: {
      fontSize: '1rem',
      lineHeight: 1.5,
      color: theme.darkTextPrimary
    },
    body2: {
      fontSize: '0.875rem',
      lineHeight: 22 / 14,
      color: theme.darkTextPrimary
    },
    caption: {
      fontSize: '0.75rem',
      lineHeight: 1.5,
      color: theme.darkTextSecondary
    },
    overline: {
      fontSize: '0.75rem',
      fontWeight: 700,
      lineHeight: 1.5,
      textTransform: 'uppercase',
      color: theme.darkTextSecondary
    },
    button: {
      textTransform: 'unset',
      fontWeight: 700,
      fontSize: '0.875rem',
      lineHeight: 24 / 14
    },
    customInput: {
      marginTop: 1,
      marginBottom: 1,
      '& > label': {
        top: 23,
        left: 0,
        color: theme.grey500,
        '&[data-shrink="false"]': {
          top: 5
        }
      },
      '& > div > input': {
        padding: '30.5px 14px 11.5px !important'
      },
      '& legend': {
        display: 'none'
      },
      '& fieldset': {
        top: 0
      }
    },
    otherInput: {
      marginTop: 1,
      marginBottom: 1
    },
    mainContent: {
      backgroundColor: theme.background,
      width: '100%',
      minHeight: 'calc(100vh - 88px)',
      flexGrow: 1,
      padding: '16px',
      paddingBottom: '30px',
      marginTop: '88px',
      marginRight: '0',
      marginBottom: '20px',
      borderRadius: '0'
    },
    menuCaption: {
      fontSize: '0.875rem',
      fontWeight: 600,
      color: theme.heading,
      padding: '6px',
      textTransform: 'capitalize',
      marginTop: '8px'
    },
    subMenuCaption: {
      fontSize: '0.75rem',
      fontWeight: 500,
      color: theme.darkTextSecondary,
      textTransform: 'capitalize'
    },
    commonAvatar: {
      cursor: 'pointer',
      borderRadius: '4px'
    },
    smallAvatar: {
      width: '24px',
      height: '24px',
      fontSize: '0.875rem'
    },
    mediumAvatar: {
      width: '40px',
      height: '40px',
      fontSize: '1.2rem'
    },
    largeAvatar: {
      width: '40px',
      height: '40px',
      fontSize: '1.25rem'
    },
    menuButton: {
      color: theme.menuButtonColor,
      background: theme.menuButton
    },
    menuChip: {
      background: theme.menuChip
    },
    CardWrapper: {
      backgroundColor: theme.mode === 'dark' ? theme.colors.darkLevel2 : theme.colors.primaryDark
    },
    SubCard: {
      border: theme.mode === 'dark' ? '1px solid rgba(227, 232, 239, 0.2)' : '1px solid rgb(227, 232, 239)'
    },
    LoginButton: {
      color: theme.darkTextPrimary,
      backgroundColor: theme.mode === 'dark' ? theme.backgroundDefault : theme.colors?.grey50,
      borderColor: theme.mode === 'dark' ? theme.colors?.grey700 : theme.colors?.grey100
    }
  };
}
