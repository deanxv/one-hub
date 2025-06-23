import { useState } from 'react'
import {
  Alert,
  Autocomplete,
  Box,
  Button,
  Checkbox,
  FormControlLabel,
  Grid,
  IconButton,
  InputAdornment,
  TextField,
  Typography
} from '@mui/material'
import { gridSpacing } from 'store/constant'
import { IconSearch, IconUserPlus } from '@tabler/icons-react'
import { fetchChannelData } from '../index'
import { API } from 'utils/api'
import { showError, showSuccess } from 'utils/common'
import { useTranslation } from 'react-i18next'

const BatchAddUserGroup = ({ groupOptions }) => {
  const [searchKeyword, setSearchKeyword] = useState('')
  const [data, setData] = useState([])
  const [selected, setSelected] = useState([])
  const [selectedUserGroup, setSelectedUserGroup] = useState('')
  const [loading, setLoading] = useState(false)
  const { t } = useTranslation()

  const handleSearch = async() => {
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

  const handleSubmit = async() => {
    if (selected.length === 0) {
      showError(t('channel_index.pleaseSelectChannelsForUserGroup'))
      return
    }

    if (!selectedUserGroup) {
      showError(t('channel_index.pleaseSelectUserGroup'))
      return
    }

    setLoading(true)
    try {
      const res = await API.put(`/api/channel/batch/add_user_group`, {
        ids: selected,
        value: selectedUserGroup
      })

      const { success, message, data } = res.data
      if (success) {
        showSuccess(t('channel_index.batchAddUserGroupSuccess', { count: data, group: selectedUserGroup }))
        // 清空选择
        setSelected([])
        setSelectedUserGroup('')
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

  const isUserGroupAlreadyExists = (channel, groupToAdd) => {
    if (!channel.group || !groupToAdd) return false
    const groups = channel.group.split(',').map((g) => g.trim())
    return groups.includes(groupToAdd)
  }

  return (
    <Grid container spacing={gridSpacing}>
      <Grid item xs={12}>
        <Alert severity="info">{t('channel_index.batchAddUserGroupTip')}</Alert>
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
          onKeyPress={(e) => {
            if (e.key === 'Enter') {
              handleSearch()
            }
          }}
          InputProps={{
            endAdornment: (
              <InputAdornment position="end">
                <IconButton aria-label={t('channel_index.searchChannelLabel')} onClick={handleSearch} edge="end">
                  <IconSearch/>
                </IconButton>
              </InputAdornment>
            )
          }}
          sx={{ '& .MuiInputBase-root': { height: '48px' } }}
        />
      </Grid>

      {data.length === 0 ? (
        <Grid item xs={12}>
          <Typography variant="body2" color="text.secondary" align="center">
            {searchKeyword ? t('channel_index.noMatchingChannels') : t('channel_index.clickSearchToGetChannels')}
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
              const hasGroup = selectedUserGroup && isUserGroupAlreadyExists(item, selectedUserGroup)
              return (
                <FormControlLabel
                  key={item.id}
                  control={<Checkbox checked={selected.includes(item.id)} onChange={() => handleSelect(item.id)}
                                     disabled={hasGroup}/>}
                  label={
                    <Box sx={{ overflow: 'hidden', minWidth: 0, flex: 1 }}>
                      <Typography
                        variant="body2"
                        sx={{
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                          width: '100%'
                        }}
                        title={item.name}
                      >
                        {item.name}
                        {hasGroup && (
                          <Typography component="span" variant="caption" color="warning.main" sx={{ ml: 1 }}>
                            {t('channel_index.channelAlreadyHasGroup')}
                          </Typography>
                        )}
                      </Typography>
                      <Typography
                        variant="caption"
                        color="text.secondary"
                        sx={{
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                          width: '100%'
                        }}
                        title={`${t('channel_index.currentGroup')}: ${item.group || t('channel_index.noGroup')}`}
                      >
                        {t('channel_index.currentGroup')}: {item.group || t('channel_index.noGroup')}
                      </Typography>
                    </Box>
                  }
                  sx={{
                    display: 'flex',
                    width: '100%',
                    opacity: hasGroup ? 0.6 : 1,
                    alignItems: 'flex-start',
                    mb: 1
                  }}
                />
              )
            })}
          </Grid>

          <Grid item xs={12}>
            <Autocomplete
              options={groupOptions}
              value={selectedUserGroup}
              onChange={(event, newValue) => {
                setSelectedUserGroup(newValue)
              }}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label={t('channel_index.selectUserGroupToAdd')}
                  placeholder={t('channel_index.pleaseSelectUserGroup')}
                />
              )}
              sx={{ mb: 2 }}
            />
          </Grid>

          <Grid item xs={12}>
            <Button
              variant="contained"
              onClick={handleSubmit}
              disabled={loading || selected.length === 0 || !selectedUserGroup}
              startIcon={<IconUserPlus/>}
              fullWidth
            >
              {loading ? t('channel_index.addingUserGroup') : t('channel_index.addUserGroupToChannels', { count: selected.length })}
            </Button>
          </Grid>
        </>
      )}
    </Grid>
  )
}

export default BatchAddUserGroup
