import { useState, useEffect } from 'react'
import {
  Alert,
  Autocomplete,
  Box,
  Button,
  Checkbox,
  Chip,
  CircularProgress,
  FormControlLabel,
  Grid,
  IconButton,
  InputAdornment,
  TextField,
  Typography,
  useMediaQuery,
  useTheme
} from '@mui/material'
import { gridSpacing } from 'store/constant'
import { IconSearch, IconPlus, IconChevronDown, IconChevronUp } from '@tabler/icons-react'
import { fetchChannelData } from '../index'
import { API } from 'utils/api'
import { copy, showError, showSuccess } from 'utils/common'
import { useTranslation } from 'react-i18next'
import { createFilterOptions } from '@mui/material/Autocomplete'
import CheckBoxOutlineBlankIcon from '@mui/icons-material/CheckBoxOutlineBlank'
import CheckBoxIcon from '@mui/icons-material/CheckBox'

const icon = <CheckBoxOutlineBlankIcon fontSize="small" />
const checkedIcon = <CheckBoxIcon fontSize="small" />
const filter = createFilterOptions()

const BatchAddModel = ({ modelOptions }) => {
  const [searchKeyword, setSearchKeyword] = useState('')
  const [data, setData] = useState([])
  const [selected, setSelected] = useState([])
  const [selectedModels, setSelectedModels] = useState([])
  const [inputValue, setInputValue] = useState('')
  const [loading, setLoading] = useState(false)
  const [searching, setSearching] = useState(false)
  const [expandedChannels, setExpandedChannels] = useState({})
  const { t } = useTranslation()
  const theme = useTheme()
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'))

  // 组件加载时自动获取全量渠道数据
  useEffect(() => {
    handleSearch()
  }, [])

  const handleSearch = async() => {
    setSearching(true)
    try {
      if (!searchKeyword.trim()) {
        // 如果没有搜索关键词，获取所有渠道
        const data = await fetchChannelData(0, 100, {}, 'desc', 'id')
        if (data) {
          setData(data.data)
        }
        return
      }

      const data = await fetchChannelData(0, 100, { name: searchKeyword }, 'desc', 'id')
      if (data) {
        setData(data.data)
      }
    } finally {
      setSearching(false)
    }
  }

  const handleSelect = (id) => {
    setSelected((prev) => {
      if (prev.includes(id)) {
        return prev.filter((i) => i !== id)
      } else {
        return [...prev, id]
      }
    })
  }

  const handleSelectAll = () => {
    if (selected.length === data.length) {
      setSelected([])
    } else {
      setSelected(data.map((item) => item.id))
    }
  }

  // 检查渠道是否已经有某个模型
  const hasAnySelectedModel = (channel) => {
    if (!channel.models || selectedModels.length === 0) return false
    const channelModelList = channel.models.split(',').map(m => m.trim())
    return selectedModels.some(model => channelModelList.includes(model))
  }

  // 切换渠道展开/折叠状态
  const toggleChannelExpanded = (channelId) => {
    setExpandedChannels(prev => ({
      ...prev,
      [channelId]: !prev[channelId]
    }))
  }

  // 获取响应式的截断长度
  const getTruncateLength = () => {
    return isMobile ? 40 : 70
  }

  // 检查模型文本是否需要折叠（与截断长度完全一致）
  const shouldTruncateModels = (models) => {
    if (!models) return false

    // 使用与截断完全相同的长度判断，确保逻辑一致
    const maxLength = getTruncateLength()
    return models.length > maxLength
  }

  // 截断模型文本显示（响应式长度）
  const truncateModels = (models) => {
    if (!models) return t('channel_index.noModels')

    const maxLength = getTruncateLength()

    if (models.length <= maxLength) return models
    return models.substring(0, maxLength) + '...'
  }

  const handleSubmit = async() => {
    if (selected.length === 0) {
      showError(t('channel_index.pleaseSelectChannelsForModel'))
      return
    }

    if (selectedModels.length === 0) {
      showError(t('channel_index.pleaseSelectModel'))
      return
    }

    setLoading(true)
    try {
      // 将选中的模型转换为逗号分隔的字符串
      const modelsString = selectedModels.map(model =>
        typeof model === 'string' ? model : model.id
      ).join(',')

      const res = await API.put(`/api/channel/batch/add_model`, {
        ids: selected,
        value: modelsString
      })

      const { success, message, data } = res.data
      if (success) {
        showSuccess(t('channel_index.batchAddModelSuccess', { count: data, model: modelsString }))
        // 清空选择
        setSelected([])
        setSelectedModels([])
        setInputValue('')
        // 重新搜索以更新显示
        handleSearch()
      } else {
        showError(message)
      }
    } catch (error) {
      showError(error.message)
    }
    setLoading(false)
  }

  return (
    <Grid container spacing={gridSpacing}>
      <Grid item xs={12}>
        <Alert severity="info">{t('channel_index.batchAddModelTip')}</Alert>
      </Grid>

      <Grid item xs={12}>
        <TextField
          fullWidth
          size="medium"
          placeholder={t('channel_index.searchChannelPlaceholder')}
          inputProps={{ 'aria-label': t('channel_index.searchChannelLabel') }}
          value={searchKeyword}
          onChange={(e) => {
            setSearchKeyword(e.target.value)
          }}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.preventDefault()
              handleSearch()
            }
          }}
          InputProps={{
            endAdornment: (
              <InputAdornment position="end">
                {searching ? (
                  <CircularProgress size={20} />
                ) : (
                  <IconButton aria-label={t('channel_index.searchChannelLabel')} onClick={handleSearch} edge="end">
                    <IconSearch/>
                  </IconButton>
                )}
              </InputAdornment>
            )
          }}
          sx={{ '& .MuiInputBase-root': { height: '48px' } }}
        />
      </Grid>

      {data.length === 0 ? (
        <Grid item xs={12}>
          <Typography variant="body2" color="text.secondary" align="center">
            {searchKeyword ? t('channel_index.noMatchingChannels') : t('channel_index.loadingChannels')}
          </Typography>
        </Grid>
      ) : (
        <>
          <Grid item xs={12}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
              <Button onClick={handleSelectAll}>
                {selected.length === data.length ? t('channel_index.unselectAll') : t('channel_index.selectAll')}
              </Button>
              <Typography variant="body2" color="text.secondary">
                {t('channel_index.selectedChannelsCount', { selected: selected.length, total: data.length })}
              </Typography>
            </Box>
          </Grid>

          <Grid item xs={12} sx={{ maxHeight: 300, overflow: 'auto' }}>
            {data.map((item) => {
              const hasConflict = hasAnySelectedModel(item)
              const isExpanded = expandedChannels[item.id]
              const needsTruncation = shouldTruncateModels(item.models)
              const displayModels = isExpanded ? item.models : truncateModels(item.models)

              return (
                <Box key={item.id} sx={{ mb: 1, width: '100%', maxWidth: '100%' }}>
                  <Box
                    sx={{
                      display: 'flex',
                      alignItems: 'flex-start',
                      width: '100%',
                      maxWidth: '100%',
                      padding: '4px 8px',
                      borderRadius: 1,
                      boxSizing: 'border-box',
                      '&:hover': {
                        backgroundColor: 'action.hover'
                      },
                      ...(hasConflict && {
                        opacity: 0.6,
                        backgroundColor: 'action.disabledBackground'
                      })
                    }}
                  >
                    <Checkbox
                      checked={selected.includes(item.id)}
                      onChange={() => handleSelect(item.id)}
                      disabled={hasConflict}
                      sx={{
                        p: isMobile ? 0.5 : 1,
                        flexShrink: 0,
                        alignSelf: 'flex-start'
                      }}
                    />
                    <Box sx={{
                      flex: 1,
                      minWidth: 0,
                      overflow: 'hidden',
                      pr: needsTruncation ? 0 : (isMobile ? 0.5 : 1)
                    }}>
                      <Typography
                        variant={isMobile ? "body2" : "body2"}
                        sx={{
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                          width: '100%',
                          mb: isMobile ? 0.25 : 0.5,
                          fontSize: isMobile ? '0.875rem' : undefined
                        }}
                        title={item.name}
                      >
                        {item.name}
                        {hasConflict && (
                          <Typography component="span" variant="caption" color="warning.main" sx={{ ml: 1 }}>
                            {t('channel_index.channelAlreadyHasModel')}
                          </Typography>
                        )}
                      </Typography>
                      <Typography
                        variant="caption"
                        color="text.secondary"
                        sx={{
                          display: 'block',
                          overflow: 'hidden',
                          textOverflow: isExpanded ? 'unset' : 'ellipsis',
                          whiteSpace: isExpanded ? 'normal' : 'nowrap',
                          wordBreak: isExpanded ? 'break-all' : 'normal',
                          lineHeight: isMobile ? 1.3 : 1.4,
                          fontSize: isMobile ? '0.75rem' : undefined
                        }}
                        title={item.models || t('channel_index.noModels')}
                      >
                        {t('channel_index.currentModels')}: {displayModels}
                      </Typography>
                    </Box>
                    {needsTruncation && (
                      <IconButton
                        size="small"
                        onClick={(e) => {
                          e.stopPropagation()
                          e.preventDefault()
                          toggleChannelExpanded(item.id)
                        }}
                        sx={{
                          p: isMobile ? 0.25 : 0.5,
                          ml: isMobile ? 0.25 : 0.5,
                          flexShrink: 0,
                          alignSelf: 'flex-start'
                        }}
                      >
                        {isExpanded ? <IconChevronUp size={16} /> : <IconChevronDown size={16} />}
                      </IconButton>
                    )}
                  </Box>
                </Box>
              )
            })}
          </Grid>

          <Grid item xs={12}>
            <Autocomplete
              multiple
              freeSolo
              disableCloseOnSelect
              options={modelOptions}
              value={selectedModels}
              inputValue={inputValue}
              onInputChange={(event, newInputValue) => {
                if (newInputValue.includes(',')) {
                  const modelsList = newInputValue
                    .split(',')
                    .map((item) => ({
                      id: item.trim(),
                      group: t('channel_edit.customModelTip')
                    }))
                    .filter((item) => item.id)

                  const updatedModels = [...new Set([...selectedModels, ...modelsList])]
                  setSelectedModels(updatedModels)
                  setInputValue('')
                } else {
                  setInputValue(newInputValue)
                }
              }}
              onChange={(e, value) => {
                setSelectedModels(value.map((item) =>
                  typeof item === 'string' ? { id: item, group: t('channel_edit.customModelTip') } : item
                ))
              }}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label={t('channel_index.selectModelToAdd')}
                  placeholder={t('channel_index.pleaseSelectModel')}
                />
              )}
              groupBy={(option) => option.group}
              getOptionLabel={(option) => {
                if (typeof option === 'string') {
                  return option
                }
                if (option.inputValue) {
                  return option.inputValue
                }
                return option.id
              }}
              filterOptions={(options, params) => {
                const filtered = filter(options, params)
                const { inputValue } = params
                const isExisting = options.some((option) => inputValue === option.id)
                if (inputValue !== '' && !isExisting) {
                  filtered.push({
                    id: inputValue,
                    group: t('channel_edit.customModelTip')
                  })
                }
                return filtered
              }}
              renderOption={(props, option, { selected }) => (
                <li {...props}>
                  <Checkbox icon={icon} checkedIcon={checkedIcon} style={{ marginRight: 8 }}
                            checked={selected}/>
                  {option.id}
                </li>
              )}
              renderTags={(value, getTagProps) =>
                value.map((option, index) => {
                  const tagProps = getTagProps({ index })
                  return (
                    <Chip
                      key={index}
                      label={option.id}
                      {...tagProps}
                      onClick={() => copy(option.id)}
                      sx={{
                        maxWidth: '100%',
                        height: 'auto',
                        margin: '3px',
                        '& .MuiChip-label': {
                          whiteSpace: 'normal',
                          wordBreak: 'break-word',
                          padding: '6px 8px',
                          lineHeight: 1.4,
                          fontWeight: 400
                        },
                        '& .MuiChip-deleteIcon': {
                          margin: '0 5px 0 -6px'
                        }
                      }}
                    />
                  )
                })
              }
              sx={{
                '& .MuiAutocomplete-tag': {
                  margin: '2px'
                },
                '& .MuiAutocomplete-inputRoot': {
                  flexWrap: 'wrap'
                },
                mb: 2
              }}
            />
          </Grid>

          <Grid item xs={12}>
            <Button
              variant="contained"
              onClick={handleSubmit}
              disabled={loading || selected.length === 0 || selectedModels.length === 0}
              startIcon={<IconPlus/>}
              fullWidth
            >
              {loading ? t('channel_index.addingModel') : t('channel_index.addModelToChannels', { count: selected.length })}
            </Button>
          </Grid>
        </>
      )}
    </Grid>
  )
}

export default BatchAddModel
