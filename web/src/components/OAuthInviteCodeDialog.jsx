import React, { useState } from 'react'
import PropTypes from 'prop-types'
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  TextField,
  Typography
} from '@mui/material'
import { useTranslation } from 'react-i18next'
import { API } from 'utils/api'

const OAuthInviteCodeDialog = ({
  open,
  onClose,
  onConfirm,
  provider
}) => {
  const { t } = useTranslation()
  const [inviteCode, setInviteCode] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async() => {
    if (!inviteCode.trim()) {
      setError('请输入邀请码')
      return
    }

    setLoading(true)
    setError('')

    try {
      const response = await API.post('/api/oauth/invite_code', {
        invite_code: inviteCode.trim()
      })

      const { success, message } = response.data
      if (success) {
        onConfirm()
      } else {
        setError(message || '验证失败')
      }
    } catch (err) {
      setError('网络错误，请重试')
    } finally {
      setLoading(false)
    }
  }

  const getProviderName = (provider) => {
    const providerNames = {
      github: 'GitHub',
      oidc: 'OIDC',
      lark: '飞书',
      linuxdo: 'Linux Do',
      wechat: '微信'
    }
    return providerNames[provider] || provider
  }

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      disableEscapeKeyDown
    >
      <DialogTitle>
        <Typography variant="h6" component="div">
          {getProviderName(provider)} 登录
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
          管理员开启了邀请码注册，请输入邀请码后继续登录
        </Typography>
      </DialogTitle>

      <Divider/>

      <DialogContent>
        <Box sx={{ pt: 2 }}>
          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          <TextField
            fullWidth
            label="邀请码"
            value={inviteCode}
            onChange={(e) => {
              setInviteCode(e.target.value)
              setError('')
            }}
            placeholder="请输入邀请码"
            margin="normal"
            required
            disabled={loading}
            helperText="必填项"
            autoFocus
          />
        </Box>
      </DialogContent>

      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button
          onClick={onClose}
          disabled={loading}
          color="inherit"
        >
          取消
        </Button>

        <Button
          onClick={handleSubmit}
          disabled={loading || !inviteCode.trim()}
          variant="contained"
          startIcon={loading ? <CircularProgress size={20}/> : null}
        >
          {loading ? '验证中...' : '确认并继续'}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

OAuthInviteCodeDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onConfirm: PropTypes.func.isRequired,
  provider: PropTypes.string.isRequired
}

export default OAuthInviteCodeDialog
