import PropTypes from 'prop-types'
import { useState } from 'react'
import { Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, Divider, Tab, Tabs } from '@mui/material'
import BatchAzureAPI from './BatchAzureAPI'
import BatchDelModel from './BatchDelModel'
import BatchAddUserGroup from './BatchAddUserGroup'
import BatchAddModel from './BatchAddModel'
import { useTranslation } from 'react-i18next'

function CustomTabPanel(props) {
  const { children, value, index, ...other } = props

  return (
    <div role="tabpanel" hidden={value !== index} id={`setting-tabpanel-${index}`}
         aria-labelledby={`channel-tab-${index}`} {...other}>
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  )
}

CustomTabPanel.propTypes = {
  children: PropTypes.node,
  index: PropTypes.number.isRequired,
  value: PropTypes.number.isRequired
}

function a11yProps(index) {
  return {
    id: `channel-tab-${index}`,
    'aria-controls': `channel-tabpanel-${index}`
  }
}

const BatchModal = ({ open, setOpen, groupOptions, modelOptions }) => {
  const { t } = useTranslation()
  const [value, setValue] = useState(0)
  const handleChange = (event, newValue) => {
    setValue(newValue)
  }

  return (
    <Dialog open={open} onClose={() => setOpen(!open)} fullWidth maxWidth={'md'}>
      <DialogTitle>
        <Box>
          <Tabs
            value={value}
            onChange={handleChange}
            aria-label="basic tabs channel"
            variant="scrollable"
            scrollButtons="auto"
            allowScrollButtonsMobile
          >
            <Tab label={t('channel_index.batchAddUserGroup')} {...a11yProps(0)} />
            <Tab label={t('channel_index.batchAddModel')} {...a11yProps(1)} />
            <Tab label={t('channel_index.batchDelete')} {...a11yProps(2)} />
            <Tab label={t('channel_index.AzureApiVersion')} {...a11yProps(3)} />
          </Tabs>
        </Box>
      </DialogTitle>
      <Divider/>
      <DialogContent>
        <CustomTabPanel value={value} index={0}>
          <BatchAddUserGroup groupOptions={groupOptions}/>
        </CustomTabPanel>
        <CustomTabPanel value={value} index={1}>
          <BatchAddModel modelOptions={modelOptions}/>
        </CustomTabPanel>
        <CustomTabPanel value={value} index={2}>
          <BatchDelModel/>
        </CustomTabPanel>
        <CustomTabPanel value={value} index={3}>
          <BatchAzureAPI/>
        </CustomTabPanel>
        <DialogActions>
          <Button onClick={() => setOpen(!open)}>{t('common.cancel')}</Button>
        </DialogActions>
      </DialogContent>
    </Dialog>
  )
}

export default BatchModal

BatchModal.propTypes = {
  open: PropTypes.bool,
  setOpen: PropTypes.func,
  groupOptions: PropTypes.array,
  modelOptions: PropTypes.array
}
