import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import { showError } from 'utils/common';
import useLogin from 'hooks/useLogin';
import { useSelector } from 'react-redux';
import { onLinuxDoOAuthClicked } from 'utils/common';

// material-ui
import { useTheme } from '@mui/material/styles';
import { Grid, Stack, Typography, useMediaQuery, CircularProgress } from '@mui/material';

// project imports
import AuthWrapper from '../AuthWrapper';
import AuthCardWrapper from '../AuthCardWrapper';
import Logo from 'ui-component/Logo';
import { useTranslation } from 'react-i18next';
import OAuthInviteCodeDialog from 'components/OAuthInviteCodeDialog';

// assets

// ================================|| AUTH3 - LOGIN ||================================ //

const LinuxDoOAuth = () => {
  const { t } = useTranslation();
  const theme = useTheme();
  const matchDownSM = useMediaQuery(theme.breakpoints.down('md'));

  const [searchParams] = useSearchParams();
  const [prompt, setPrompt] = useState(t('common.processing'));
  const [showInviteDialog, setShowInviteDialog] = useState(false);
  const { linuxDoLogin } = useLogin();
  const siteInfo = useSelector((state) => state.siteInfo);

  let navigate = useNavigate();

  const sendCode = async (code, state, count) => {
    const { success, message } = await linuxDoLogin(code, state);
    if (!success) {
      // 检查是否需要邀请码
      if (message && message.startsWith('NEED_INVITE_CODE:')) {
        const actualMessage = message.substring('NEED_INVITE_CODE:'.length);
        setPrompt(actualMessage || '需要邀请码');
        setShowInviteDialog(true);
        return;
      }

      if (message) {
        showError(message);
      }
      if (count === 0) {
        setPrompt(t('login.linuxDoError'));
        await new Promise((resolve) => setTimeout(resolve, 2000));
        navigate('/login');
        return;
      }
      count++;
      setPrompt(t('login.linuxDoCountError', { count }));
      await new Promise((resolve) => setTimeout(resolve, 2000));
      await sendCode(code, state, count);
    }
  };

  // 处理邀请码确认 - 重新授权
  const handleInviteCodeConfirm = () => {
    setShowInviteDialog(false);
    setPrompt('邀请码已设置，正在重新授权...');

    // 使用正确的OAuth发起流程，重新生成state并跳转到LinuxDo
    setTimeout(() => {
      onLinuxDoOAuthClicked(siteInfo.linuxDo_client_id);
    }, 1000);
  };

  // 处理邀请码对话框关闭
  const handleInviteCodeClose = () => {
    setShowInviteDialog(false);
    navigate('/login');
  };

  useEffect(() => {
    let code = searchParams.get('code');
    let state = searchParams.get('state');
    sendCode(code, state, 0).then();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <>
      <AuthWrapper>
        <Grid container direction="column" justifyContent="flex-end">
          <Grid item xs={12}>
            <Grid container justifyContent="center" alignItems="center" sx={{ minHeight: 'calc(100vh - 136px)' }}>
              <Grid item sx={{ m: { xs: 1, sm: 3 }, mb: 0 }}>
                <AuthCardWrapper>
                  <Grid container spacing={2} alignItems="center" justifyContent="center">
                    <Grid item sx={{ mb: 3 }}>
                      <Link to="#">
                        <Logo />
                      </Link>
                    </Grid>
                    <Grid item xs={12}>
                      <Grid container direction={matchDownSM ? 'column-reverse' : 'row'} alignItems="center" justifyContent="center">
                        <Grid item>
                          <Stack alignItems="center" justifyContent="center" spacing={1}>
                            <Typography color={theme.palette.primary.main} gutterBottom variant={matchDownSM ? 'h3' : 'h2'}>
                              {t('login.linuxDoLogin')}
                            </Typography>
                          </Stack>
                        </Grid>
                      </Grid>
                    </Grid>
                    <Grid item xs={12} container direction="column" justifyContent="center" alignItems="center" style={{ height: '200px' }}>
                      <CircularProgress />
                      <Typography variant="h3" paddingTop={'20px'}>
                        {prompt}
                      </Typography>
                    </Grid>
                  </Grid>
                </AuthCardWrapper>
              </Grid>
            </Grid>
          </Grid>
        </Grid>
      </AuthWrapper>

      {/* 邀请码对话框 */}
      <OAuthInviteCodeDialog
        open={showInviteDialog}
        onClose={handleInviteCodeClose}
        onConfirm={handleInviteCodeConfirm}
        provider="linuxdo"
      />
    </>
  );
};

export default LinuxDoOAuth;
