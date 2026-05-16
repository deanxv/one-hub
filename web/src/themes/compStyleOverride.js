import { varAlpha } from './utils';

export default function componentStyleOverrides(theme) {
  const isDark = theme.mode === 'dark';

  return {
    MuiCssBaseline: {
      styleOverrides: `
        * {
          box-sizing: border-box;
        }
        html {
          height: 100%;
          -webkit-overflow-scrolling: touch;
        }
        html, body {
          font-family: "Public Sans Variable", -apple-system, BlinkMacSystemFont, sans-serif;
          -webkit-font-smoothing: antialiased;
          -moz-osx-font-smoothing: grayscale;
        }
        body, #root, #root__layout {
          display: flex;
          flex: 1 1 auto;
          min-height: 100%;
          flex-direction: column;
        }
        img {
          max-width: 100%;
          vertical-align: middle;
        }
        ul {
          margin: 0;
          padding: 0;
          list-style-type: none;
        }
        input[type='number'] {
          -moz-appearance: textfield;
          appearance: none;
        }
        input[type='number']::-webkit-outer-spin-button,
        input[type='number']::-webkit-inner-spin-button {
          margin: 0;
          -webkit-appearance: none;
        }
        .apexcharts-title-text {
          fill: ${theme.textDark} !important
        }
        .apexcharts-text {
          fill: ${theme.textDark} !important
        }
        .apexcharts-legend-text {
          color: ${theme.textDark} !important
        }
        .apexcharts-menu {
          background: ${theme.backgroundDefault} !important
        }
        .apexcharts-gridline, .apexcharts-xaxistooltip-background, .apexcharts-yaxistooltip-background {
          stroke: ${theme.divider} !important;
        }
      `
    },
    MuiButtonBase: {
      styleOverrides: {
        root: {
          fontFamily: '"Public Sans Variable", -apple-system, BlinkMacSystemFont, sans-serif'
        }
      }
    },
    MuiButton: {
      defaultProps: {
        color: 'inherit',
        disableElevation: true
      },
      styleOverrides: {
        root: {
          fontWeight: 700,
          borderRadius: '8px',
          textTransform: 'unset',
          boxShadow: 'none',
          '&.Mui-disabled': {
            color: theme.colors?.grey500
          },
          '&:hover': {
            boxShadow: 'none'
          }
        },
        contained: {
          color: isDark ? theme.colors?.grey800 : '#FFFFFF',
          backgroundColor: isDark ? '#FFFFFF' : theme.colors?.grey800,
          '&:hover': {
            backgroundColor: isDark ? theme.colors?.grey300 : theme.colors?.grey700
          }
        },
        containedPrimary: {
          backgroundColor: theme.colors?.primaryMain,
          '&:hover': {
            backgroundColor: theme.colors?.primaryDark,
            boxShadow: `0 8px 16px 0 ${varAlpha(theme.colors?.primaryMain, 0.24)}`
          }
        },
        containedSecondary: {
          '&:hover': {
            boxShadow: `0 8px 16px 0 ${varAlpha(theme.colors?.secondaryMain, 0.24)}`
          }
        },
        containedError: {
          '&:hover': {
            boxShadow: `0 8px 16px 0 ${varAlpha(theme.colors?.errorMain, 0.24)}`
          }
        },
        outlined: {
          '&:hover': {
            borderColor: 'currentColor',
            boxShadow: '0 0 0 0.75px currentColor'
          }
        },
        outlinedInherit: {
          borderColor: isDark ? varAlpha(theme.colors?.grey500, 0.32) : varAlpha(theme.colors?.grey500, 0.32),
          '&:hover': {
            backgroundColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.04)'
          }
        },
        outlinedPrimary: {
          borderColor: varAlpha(theme.colors?.primaryMain, 0.48)
        },
        text: {
          '&:hover': {
            backgroundColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.04)'
          }
        },
        sizeSmall: {
          height: 34,
          fontSize: '0.8125rem',
          paddingLeft: '12px',
          paddingRight: '12px'
        },
        sizeMedium: {
          paddingLeft: '12px',
          paddingRight: '12px'
        },
        sizeLarge: {
          height: 48,
          paddingLeft: '16px',
          paddingRight: '16px'
        }
      }
    },
    MuiIconButton: {
      styleOverrides: {
        root: {
          borderRadius: '8px',
          padding: '8px',
          color: theme.darkTextPrimary,
          '&:hover': {
            backgroundColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.04)'
          }
        },
        sizeSmall: {
          padding: '4px'
        }
      }
    },
    MuiPaper: {
      defaultProps: {
        elevation: 0
      },
      styleOverrides: {
        root: {
          backgroundImage: 'none'
        },
        outlined: {
          borderColor: varAlpha(theme.colors?.grey500, 0.16)
        },
        rounded: {
          borderRadius: `${theme?.customization?.borderRadius || 8}px`
        }
      }
    },
    MuiCard: {
      styleOverrides: {
        root: {
          position: 'relative',
          borderRadius: `${(theme?.customization?.borderRadius || 8) * 2}px`,
          padding: 0,
          boxShadow: isDark
            ? `0 0 2px 0 ${varAlpha(theme.colors?.grey500, 0.2)}, 0 12px 24px -4px ${varAlpha(theme.colors?.grey500, 0.12)}`
            : `0 0 2px 0 ${varAlpha(theme.colors?.grey500, 0.2)}, 0 12px 24px -4px ${varAlpha(theme.colors?.grey500, 0.12)}`,
          zIndex: 0,
          overflow: 'hidden',
          '& .MuiTableContainer-root': {
            borderRadius: 0
          }
        }
      }
    },
    MuiCardHeader: {
      defaultProps: {
        titleTypographyProps: { variant: 'h6' },
        subheaderTypographyProps: { variant: 'body2', marginTop: '4px' }
      },
      styleOverrides: {
        root: {
          padding: '24px 24px 0'
        }
      }
    },
    MuiCardContent: {
      styleOverrides: {
        root: {
          padding: '24px',
          '&:last-child': {
            paddingBottom: '24px'
          },
          '& .MuiTableContainer-root': {
            margin: '-24px',
            width: 'calc(100% + 48px)',
            maxWidth: 'calc(100% + 48px)'
          }
        }
      }
    },
    MuiCardActions: {
      styleOverrides: {
        root: {
          padding: '16px 20px'
        }
      }
    },
    MuiAutocomplete: {
      styleOverrides: {
        popper: {
          borderRadius: '8px',
          color: theme.darkTextPrimary
        },
        listbox: {
          padding: '4px 0'
        },
        option: {
          fontSize: '0.875rem',
          fontWeight: 400,
          lineHeight: '1.43',
          padding: '6px 8px',
          borderRadius: `${(theme?.customization?.borderRadius || 8) * 0.75}px`,
          '&:not(:last-of-type)': {
            marginBottom: 4
          },
          '&:hover': {
            backgroundColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.04)'
          },
          '&[aria-selected="true"]': {
            fontWeight: 600,
            backgroundColor: varAlpha(theme.colors?.grey500, 0.16),
            '&:hover': {
              backgroundColor: varAlpha(theme.colors?.grey500, 0.08)
            }
          }
        }
      }
    },
    MuiListItemButton: {
      styleOverrides: {
        root: {
          color: theme.darkTextPrimary,
          borderRadius: '8px',
          padding: '8px 16px',
          '&.Mui-selected': {
            color: theme.menuSelected,
            backgroundColor: theme.menuSelectedBack,
            '&:hover': {
              backgroundColor: theme.menuSelectedBack
            },
            '& .MuiListItemIcon-root': {
              color: theme.menuSelected
            }
          },
          '&:hover': {
            backgroundColor: isDark ? 'rgba(255, 255, 255, 0.04)' : 'rgba(0, 0, 0, 0.04)',
            color: theme.menuSelected,
            '& .MuiListItemIcon-root': {
              color: theme.menuSelected
            }
          }
        }
      }
    },
    MuiListItemIcon: {
      styleOverrides: {
        root: {
          color: 'inherit',
          minWidth: 'auto',
          marginRight: '16px'
        }
      }
    },
    MuiListItemText: {
      defaultProps: {
        primaryTypographyProps: { typography: 'subtitle2' }
      },
      styleOverrides: {
        root: {
          margin: 0
        },
        primary: {
          color: theme.textDark,
          fontSize: '0.875rem'
        },
        secondary: {
          fontSize: '0.75rem',
          color: theme.darkTextSecondary
        },
        multiline: {
          margin: 0
        }
      }
    },
    MuiInputBase: {
      styleOverrides: {
        root: {
          fontSize: '0.875rem',
          lineHeight: '1.5',
          '&.Mui-disabled': {
            '& svg': {
              color: theme.colors?.grey500
            }
          },
          '& .MuiInputBase-input:focus': {
            borderRadius: 'inherit'
          }
        },
        input: {
          color: theme.textDark,
          fontSize: '0.9375rem',
          '&::placeholder': {
            opacity: 1,
            color: theme.colors?.grey500
          }
        }
      }
    },
    MuiOutlinedInput: {
      styleOverrides: {
        root: {
          borderRadius: `${theme?.customization?.borderRadius || 8}px`,
          '& .MuiOutlinedInput-notchedOutline': {
            borderColor: varAlpha(theme.colors?.grey500, 0.2),
            borderWidth: '1px',
            transition: 'border-color 0.2s ease-in-out'
          },
          '&:hover .MuiOutlinedInput-notchedOutline': {
            borderColor: theme.darkTextPrimary,
            borderWidth: '1px'
          },
          '&.Mui-focused': {
            '& .MuiOutlinedInput-notchedOutline': {
              borderColor: theme.darkTextPrimary,
              borderWidth: '2px'
            }
          },
          '&.Mui-error': {
            '& .MuiOutlinedInput-notchedOutline': {
              borderColor: theme.colors?.errorMain
            }
          },
          '&.Mui-disabled': {
            '& .MuiOutlinedInput-notchedOutline': {
              borderColor: varAlpha(theme.colors?.grey500, 0.24)
            }
          },
          '& input::-webkit-outer-spin-button, & input::-webkit-inner-spin-button': {
            WebkitAppearance: 'none',
            margin: 0
          },
          '& input[type=number]': {
            MozAppearance: 'textfield'
          }
        },
        input: {
          padding: '14px 16px',
          height: 'auto',
          fontSize: '0.9375rem'
        },
        inputMultiline: {
          padding: '4px 8px'
        },
        sizeSmall: {
          '& input': {
            padding: '10px 14px'
          }
        }
      }
    },
    MuiFormLabel: {
      styleOverrides: {
        root: {
          fontSize: '0.875rem',
          color: theme.colors?.grey500,
          '&.MuiInputLabel-shrink': {
            fontSize: '1rem',
            fontWeight: 600,
            color: theme.darkTextSecondary,
            '&.Mui-focused': {
              color: theme.darkTextPrimary
            },
            '&.Mui-error': {
              color: theme.colors?.errorMain
            },
            '&.Mui-disabled': {
              color: theme.colors?.grey500
            }
          }
        }
      }
    },
    MuiFormHelperText: {
      defaultProps: {
        component: 'div'
      },
      styleOverrides: {
        root: {
          fontSize: '0.75rem',
          marginLeft: '4px',
          marginTop: '8px'
        }
      }
    },
    MuiFormControlLabel: {
      styleOverrides: {
        label: {
          fontSize: '0.875rem'
        }
      }
    },
    MuiInputAdornment: {
      styleOverrides: {
        root: {
          marginLeft: 0,
          width: 'auto',
          '& .MuiIconButton-root': {
            padding: 0,
            width: '20px',
            height: '20px',
            minWidth: '20px',
            margin: 0
          },
          '& .MuiSvgIcon-root, & .iconify': {
            fontSize: '16px',
            width: '16px',
            height: '16px'
          }
        },
        positionEnd: {
          marginLeft: 0,
          paddingLeft: 0
        }
      }
    },
    MuiSlider: {
      defaultProps: {
        size: 'small'
      },
      styleOverrides: {
        root: {
          height: 6,
          '&.Mui-disabled': {
            color: varAlpha(theme.colors?.grey500, 0.48)
          }
        },
        rail: {
          opacity: 0.12,
          height: 6,
          backgroundColor: theme.colors?.grey500
        },
        track: {
          height: 6
        },
        mark: {
          width: 1,
          height: 4,
          backgroundColor: varAlpha(theme.colors?.grey500, 0.48)
        },
        valueLabel: {
          borderRadius: 8,
          backgroundColor: isDark ? theme.colors?.grey700 : theme.colors?.grey800
        },
        thumb: {
          width: 16,
          height: 16,
          borderWidth: 1,
          borderStyle: 'solid',
          borderColor: varAlpha(theme.colors?.grey500, 0.08),
          backgroundColor: '#FFFFFF'
        },
        sizeSmall: {
          '& .MuiSlider-thumb': {
            width: 16,
            height: 16
          },
          '& .MuiSlider-rail': { height: 6 },
          '& .MuiSlider-track': { height: 6 },
          '& .MuiSlider-mark': { height: 4 }
        }
      }
    },
    MuiDivider: {
      styleOverrides: {
        root: {
          borderColor: theme.divider,
          opacity: 1
        }
      }
    },
    MuiAvatar: {
      styleOverrides: {
        root: {
          fontSize: '1rem',
          fontWeight: 600
        },
        rounded: {
          borderRadius: `${(theme?.customization?.borderRadius || 8) * 1.5}px`
        },
        colorDefault: {
          color: theme.darkTextSecondary,
          backgroundColor: varAlpha(theme.colors?.grey500, 0.24)
        }
      }
    },
    MuiChip: {
      styleOverrides: {
        root: {
          transition: 'all 0.2s ease-in-out',
          '&.Mui-disabled': {
            opacity: 1,
            color: varAlpha(theme.colors?.grey500, 0.8),
            backgroundColor: varAlpha(theme.colors?.grey500, 0.24)
          }
        },
        label: {
          fontWeight: 500
        },
        icon: {
          color: 'currentColor'
        },
        deleteIcon: {
          opacity: 0.48,
          color: 'currentColor',
          '&:hover': {
            opacity: 1,
            color: 'currentColor'
          }
        },
        sizeMedium: {
          borderRadius: `${(theme?.customization?.borderRadius || 8) * 1.25}px`
        },
        sizeSmall: {
          borderRadius: `${theme?.customization?.borderRadius || 8}px`
        },
        filled: {
          color: '#FFFFFF',
          backgroundColor: theme.darkTextPrimary,
          '&:hover': {
            backgroundColor: isDark ? theme.colors?.grey100 : theme.colors?.grey700
          }
        },
        outlined: {
          borderColor: varAlpha(theme.colors?.grey500, 0.32)
        }
      }
    },
    MuiTableContainer: {
      styleOverrides: {
        root: {
          position: 'relative',
          scrollbarWidth: 'thin',
          overflowX: 'auto',
          overflowY: 'auto',
          borderRadius: `${theme?.customization?.borderRadius || 8}px`,
          boxShadow: 'none'
        }
      }
    },
    MuiTable: {
      styleOverrides: {
        root: {
          borderCollapse: 'separate',
          borderSpacing: 0,
          width: '100%',
          margin: 0,
          padding: 0
        }
      }
    },
    MuiTableHead: {
      styleOverrides: {
        root: {
          backgroundColor: isDark ? '#28323D' : theme.colors?.grey200,
          width: '100%',
          margin: 0,
          '& tr': {
            width: '100%',
            '& th:first-of-type': {
              borderTopLeftRadius: 0
            },
            '& th:last-of-type': {
              borderTopRightRadius: 0
            }
          }
        }
      }
    },
    MuiTableCell: {
      styleOverrides: {
        root: {
          borderBottomStyle: 'dashed',
          borderBottomColor: theme.divider,
          fontSize: '0.875rem',
          padding: '16px 12px',
          textAlign: 'center',
          '&:first-of-type': {
            paddingLeft: '12px'
          },
          '&:last-of-type': {
            paddingRight: '12px'
          }
        },
        head: {
          fontSize: 14,
          fontWeight: 600,
          color: theme.darkTextSecondary,
          backgroundColor: isDark ? '#28323D' : theme.colors?.grey200,
          borderBottom: 'none',
          whiteSpace: 'nowrap',
          padding: '14px 12px',
          textAlign: 'center'
        },
        body: {
          color: theme.textDark
        }
      }
    },
    MuiTableRow: {
      styleOverrides: {
        root: {
          transition: 'background-color 0.2s ease',
          '&.Mui-selected': {
            backgroundColor: varAlpha(theme.colors?.primaryDark, 0.04),
            '&:hover': {
              backgroundColor: varAlpha(theme.colors?.primaryDark, 0.08)
            }
          },
          '&:last-of-type': {
            '& .MuiTableCell-root': {
              borderColor: 'transparent'
            }
          }
        }
      }
    },
    MuiTablePagination: {
      defaultProps: {
        backIconButtonProps: { size: 'small' },
        nextIconButtonProps: { size: 'small' }
      },
      styleOverrides: {
        root: {
          width: '100%',
          color: theme.textDark,
          borderTop: `1px dashed ${theme.divider}`,
          overflow: 'auto',
          minHeight: '56px',
          margin: 0
        },
        toolbar: {
          height: 64
        },
        actions: {
          marginRight: 8
        },
        select: {
          paddingLeft: 8,
          display: 'flex',
          alignItems: 'center',
          '&:focus': {
            borderRadius: `${theme?.customization?.borderRadius || 8}px`
          }
        },
        selectIcon: {
          right: 4,
          width: 16,
          height: 16,
          top: 'calc(50% - 8px)'
        }
      }
    },
    MuiTooltip: {
      styleOverrides: {
        tooltip: {
          backgroundColor: isDark ? theme.colors?.grey700 : theme.colors?.grey800,
          borderRadius: '6px',
          fontWeight: 400,
          fontSize: '0.75rem',
          padding: '6px 10px'
        },
        arrow: {
          color: isDark ? theme.colors?.grey700 : theme.colors?.grey800
        },
        popper: {
          '&[data-popper-placement*="bottom"] .MuiTooltip-tooltip': {
            marginTop: 12
          },
          '&[data-popper-placement*="top"] .MuiTooltip-tooltip': {
            marginBottom: 12
          },
          '&[data-popper-placement*="right"] .MuiTooltip-tooltip': {
            marginLeft: 12
          },
          '&[data-popper-placement*="left"] .MuiTooltip-tooltip': {
            marginRight: 12
          }
        }
      }
    },
    MuiAlert: {
      styleOverrides: {
        root: {
          fontSize: '0.875rem',
          borderRadius: '8px',
          alignItems: 'center',
          padding: '12px 16px'
        },
        icon: {
          opacity: 1
        },
        standardSuccess: {
          backgroundColor: isDark ? theme.colors?.successDark : theme.colors?.success200,
          color: isDark ? theme.colors?.success200 : theme.colors?.successDark,
          '& .MuiAlert-icon': {
            color: isDark ? theme.colors?.successLight : theme.colors?.successMain
          }
        },
        standardError: {
          backgroundColor: isDark ? theme.colors?.errorDark : '#FFE9D5',
          color: isDark ? '#FFE9D5' : theme.colors?.errorDark,
          '& .MuiAlert-icon': {
            color: isDark ? theme.colors?.errorLight : theme.colors?.errorMain
          }
        },
        standardWarning: {
          backgroundColor: isDark ? theme.colors?.warningDark : '#FFF5CC',
          color: isDark ? '#FFF5CC' : theme.colors?.warningDark,
          '& .MuiAlert-icon': {
            color: isDark ? theme.colors?.warningLight : theme.colors?.warningMain
          }
        },
        standardInfo: {
          backgroundColor: isDark ? theme.colors?.infoDark || '#006C9C' : '#CAFDF5',
          color: isDark ? '#CAFDF5' : theme.colors?.infoDark || '#006C9C',
          '& .MuiAlert-icon': {
            color: isDark ? theme.colors?.infoLight || '#61F3F3' : theme.colors?.infoMain || '#00B8D9'
          }
        }
      }
    },
    MuiTabs: {
      defaultProps: {
        textColor: 'inherit',
        variant: 'scrollable',
        allowScrollButtonsMobile: true
      },
      styleOverrides: {
        flexContainer: {
          gap: '24px',
          '@media (min-width:600px)': {
            gap: '40px'
          }
        },
        indicator: {
          backgroundColor: 'currentColor'
        }
      }
    },
    MuiTab: {
      defaultProps: {
        disableRipple: true,
        iconPosition: 'start'
      },
      styleOverrides: {
        root: {
          opacity: 1,
          minWidth: 48,
          minHeight: 48,
          padding: '8px 0',
          textTransform: 'none',
          fontSize: '0.875rem',
          fontWeight: 500,
          color: theme.darkTextSecondary,
          '&.Mui-selected': {
            color: theme.darkTextPrimary,
            fontWeight: 600
          }
        }
      }
    },
    MuiDialog: {
      styleOverrides: {
        paper: {
          borderRadius: `${(theme?.customization?.borderRadius || 8) * 2}px`,
          boxShadow: `-40px 40px 80px -8px rgba(0, 0, 0, 0.24)`,
          overflow: 'visible',
          margin: '16px',
          '&.MuiPaper-rounded': {
            borderRadius: `${(theme?.customization?.borderRadius || 8) * 2}px`
          }
        },
        paperFullScreen: {
          borderRadius: 0
        }
      }
    },
    MuiDialogTitle: {
      styleOverrides: {
        root: {
          fontSize: '1.25rem',
          fontWeight: 600,
          padding: '24px',
          color: theme.textDark
        }
      }
    },
    MuiDialogContent: {
      styleOverrides: {
        root: {
          padding: '0 24px',
          fontSize: '0.9375rem',
          color: theme.darkTextPrimary
        },
        dividers: {
          borderTop: 0,
          borderBottomStyle: 'dashed',
          paddingBottom: '24px'
        }
      }
    },
    MuiDialogActions: {
      defaultProps: {
        disableSpacing: true
      },
      styleOverrides: {
        root: {
          padding: '24px',
          '& > :not(:first-of-type)': {
            marginLeft: '12px'
          }
        }
      }
    },
    MuiLink: {
      defaultProps: {
        underline: 'hover'
      },
      styleOverrides: {
        root: {
          color: theme.colors?.primaryMain
        }
      }
    },
    MuiBadge: {
      styleOverrides: {
        dot: {
          borderRadius: '50%'
        }
      }
    },
    MuiSelect: {
      styleOverrides: {
        icon: {
          right: 10,
          width: 18,
          height: 18,
          top: 'calc(50% - 9px)',
          color: theme.darkTextSecondary,
          transition: 'transform 0.2s ease-in-out'
        },
        iconOpen: {
          transform: 'rotate(180deg)'
        }
      }
    },
    MuiMenuItem: {
      styleOverrides: {
        root: {
          fontSize: '0.875rem',
          padding: '6px 8px',
          borderRadius: `${(theme?.customization?.borderRadius || 8) * 0.75}px`,
          '&:not(:last-of-type)': {
            marginBottom: 4
          },
          '&.Mui-selected': {
            fontWeight: 600,
            backgroundColor: varAlpha(theme.colors?.grey500, 0.16),
            '&:hover': {
              backgroundColor: varAlpha(theme.colors?.grey500, 0.08)
            }
          }
        }
      }
    },
    MuiDataGrid: {
      styleOverrides: {
        root: {
          border: 'none',
          borderRadius: 0,
          backgroundColor: isDark ? theme.paper : '#FFFFFF',
          overflow: 'hidden',
          width: '100%',
          margin: 0,
          padding: 0,
          scrollbarWidth: 'thin',
          '& .MuiPaper-root': {
            borderRadius: 0
          },
          '&.MuiPaper-root': {
            borderRadius: 0
          },
          '& .MuiDataGrid-main': {
            width: '100%',
            margin: 0,
            padding: 0,
            '& .MuiDataGrid-columnHeaders': {
              borderBottom: `1px dashed ${theme.divider}`,
              borderRadius: 0,
              backgroundColor: isDark ? '#28323D' : theme.colors?.grey200,
              minHeight: '48px',
              width: '100%',
              margin: 0
            },
            '& .MuiDataGrid-virtualScroller': {
              backgroundColor: isDark ? theme.paper : '#fff',
              width: '100%',
              margin: 0
            },
            '& .MuiDataGrid-columnHeadersInner': {
              width: '100%',
              margin: 0,
              padding: 0,
              '& .MuiDataGrid-columnHeader': {
                padding: 0,
                margin: 0,
                '& .MuiDataGrid-columnHeaderTitleContainer': {
                  justifyContent: 'center',
                  padding: '0 16px',
                  margin: 0
                },
                '&:first-of-type': {
                  '& .MuiDataGrid-columnHeaderTitleContainer': {
                    paddingLeft: '24px'
                  }
                },
                '&:last-of-type': {
                  '& .MuiDataGrid-columnHeaderTitleContainer': {
                    paddingRight: '24px'
                  }
                }
              }
            },
            '& .MuiDataGrid-cellContent': {
              justifyContent: 'center',
              width: '100%',
              display: 'flex'
            }
          },
          '& .MuiDataGrid-filler > div': {
            borderTopStyle: 'dashed'
          },
          '& .MuiDataGrid-topContainer::after': {
            height: 0
          },
          footerContainer: {
            borderTop: `1px dashed ${theme.divider}`,
            borderTopStyle: 'dashed',
            minHeight: 'auto',
            backgroundColor: isDark ? '#28323D' : theme.colors?.grey200,
            width: '100%',
            margin: 0,
            padding: '0 24px',
            '& .MuiTablePagination-root': {
              overflow: 'visible',
              backgroundColor: 'transparent',
              color: theme.textDark,
              borderTop: 'none'
            },
            '& .MuiToolbar-root': {
              minHeight: '56px',
              padding: '0',
              '& > p:first-of-type': {
                fontSize: '0.875rem',
                color: theme.darkTextSecondary
              }
            }
          }
        },
        columnHeader: {
          padding: '12px 16px',
          fontSize: 14,
          fontWeight: 600,
          color: theme.darkTextSecondary,
          height: '48px',
          textAlign: 'center',
          '&:focus': {
            outline: 'none'
          },
          '&:focus-within': {
            outline: 'none'
          },
          '&--sorted': {
            color: theme.darkTextPrimary
          }
        },
        columnHeaderTitle: {
          color: theme.darkTextSecondary,
          fontWeight: 600,
          fontSize: 14,
          textAlign: 'center'
        },
        columnSeparator: {
          color: theme.divider
        },
        cell: {
          fontSize: '0.875rem',
          padding: '12px 16px',
          borderTopStyle: 'dashed',
          borderBottom: `1px dashed ${theme.divider}`,
          textAlign: 'center',
          '&:focus': {
            outline: 'none'
          },
          '&:focus-within': {
            outline: 'none'
          },
          '&--editing': {
            boxShadow: 'none',
            backgroundColor: varAlpha(theme.colors?.primaryMain, 0.08)
          }
        },
        row: {
          transition: 'background-color 0.2s ease',
          '&:hover': {
            backgroundColor: isDark ? 'rgba(0, 0, 0, 0.3)' : 'rgba(0, 0, 0, 0.04)'
          },
          '&.Mui-selected': {
            backgroundColor: varAlpha(theme.colors?.primaryDark, 0.04),
            '&:hover': {
              backgroundColor: varAlpha(theme.colors?.primaryDark, 0.08)
            }
          }
        },
        selectedRowCount: {
          display: 'none',
          whiteSpace: 'nowrap'
        },
        toolbarContainer: {
          backgroundColor: isDark ? theme.paper : '#FFFFFF',
          gap: '16px',
          padding: '16px',
          '& .MuiButton-root': {
            marginRight: '8px'
          }
        }
      }
    }
  };
}
