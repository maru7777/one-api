import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Form, Card, Modal, Message, Divider, Image } from 'semantic-ui-react';
import { useParams, useNavigate } from 'react-router-dom';
import { API, showError, showSuccess } from '../../helpers';
import { renderQuota, renderQuotaWithPrompt } from '../../helpers/render';
import QRCode from 'qrcode';

const EditUser = () => {
  const { t } = useTranslation();
  const params = useParams();
  const userId = params.id;
  const [loading, setLoading] = useState(true);
  const [inputs, setInputs] = useState({
    username: '',
    display_name: '',
    password: '',
    github_id: '',
    wechat_id: '',
    email: '',
    quota: 0,
    group: 'default',
  });
  const [groupOptions, setGroupOptions] = useState([]);

  // TOTP related state
  const [totpEnabled, setTotpEnabled] = useState(false);
  const [showTotpSetup, setShowTotpSetup] = useState(false);
  const [totpSecret, setTotpSecret] = useState('');
  const [totpQRCode, setTotpQRCode] = useState('');
  const [totpCode, setTotpCode] = useState('');
  const [totpLoading, setTotpLoading] = useState(false);
  const {
    username,
    display_name,
    password,
    github_id,
    wechat_id,
    email,
    quota,
    group,
  } = inputs;
  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };
  const fetchGroups = async () => {
    try {
      let res = await API.get(`/api/group/`);
      setGroupOptions(
        res.data.data.map((group) => ({
          key: group,
          text: group,
          value: group,
        }))
      );
    } catch (error) {
      showError(error.message);
    }
  };
  const navigate = useNavigate();
  const handleCancel = () => {
    navigate('/setting');
  };
  const loadUser = async () => {
    let res = undefined;
    if (userId) {
      res = await API.get(`/api/user/${userId}`);
    } else {
      res = await API.get(`/api/user/self`);
    }
    const { success, message, data } = res.data;
    if (success) {
      data.password = '';
      setInputs(data);
      // For admin editing other users, set TOTP status from user data
      if (userId) {
        setTotpEnabled(data.totp_secret && data.totp_secret !== '');
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const loadTotpStatus = async () => {
    if (userId) {
      // For admin editing other users, TOTP status is loaded from user data
      return;
    }
    // Only load TOTP status API for self
    try {
      const res = await API.get('/api/user/totp/status');
      if (res.data.success) {
        setTotpEnabled(res.data.data.totp_enabled);
      }
    } catch (error) {
      console.error('Failed to load TOTP status:', error);
    }
  };

  const setupTotp = async () => {
    setTotpLoading(true);
    try {
      const res = await API.get('/api/user/totp/setup');
      if (res.data.success) {
        setTotpSecret(res.data.data.secret);
        // Generate QR code from URI
        const qrCodeDataURL = await QRCode.toDataURL(res.data.data.qr_code);
        setTotpQRCode(qrCodeDataURL);
        setShowTotpSetup(true);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError('Failed to setup TOTP');
    }
    setTotpLoading(false);
  };

  const confirmTotp = async () => {
    if (!totpCode) {
      showError('Please enter the TOTP code');
      return;
    }
    setTotpLoading(true);
    try {
      const res = await API.post('/api/user/totp/confirm', {
        totp_code: totpCode,
      });
      if (res.data.success) {
        showSuccess('TOTP has been successfully enabled');
        setTotpEnabled(true);
        setShowTotpSetup(false);
        setTotpCode('');
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError('Failed to confirm TOTP');
    }
    setTotpLoading(false);
  };

  const disableTotp = async () => {
    if (!totpCode) {
      showError('Please enter the TOTP code to disable');
      return;
    }
    setTotpLoading(true);
    try {
      const res = await API.post('/api/user/totp/disable', {
        totp_code: totpCode,
      });
      if (res.data.success) {
        showSuccess('TOTP has been successfully disabled');
        setTotpEnabled(false);
        setTotpCode('');
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError('Failed to disable TOTP');
    }
    setTotpLoading(false);
  };

  const adminDisableTotp = async () => {
    setTotpLoading(true);
    try {
      const res = await API.post(`/api/user/totp/disable/${userId}`);
      if (res.data.success) {
        showSuccess('TOTP has been successfully disabled for the user');
        setTotpEnabled(false);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError('Failed to disable TOTP');
    }
    setTotpLoading(false);
  };
  useEffect(() => {
    loadUser().then();
    loadTotpStatus().then();
    if (userId) {
      fetchGroups().then();
    }
  }, [userId]); // eslint-disable-line react-hooks/exhaustive-deps

  const submit = async () => {
    let res = undefined;
    if (userId) {
      let data = { ...inputs, id: parseInt(userId) };
      if (typeof data.quota === 'string') {
        data.quota = parseInt(data.quota);
      }
      res = await API.put(`/api/user/`, data);
    } else {
      res = await API.put(`/api/user/self`, inputs);
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('user.messages.update_success'));
    } else {
      showError(message);
    }
  };

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>{t('user.edit.title')}</Card.Header>
          {loading ? (
            <div style={{
              minHeight: '400px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              backgroundColor: 'var(--card-bg)',
              borderRadius: '8px',
              margin: '1rem 0'
            }}>
              <div style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                gap: '1rem',
                color: 'var(--text-secondary)'
              }}>
                <div className="ui active inline loader"></div>
                <span>{t('user.edit.loading')}</span>
              </div>
            </div>
          ) : (
            <Form autoComplete='new-password'>
            <Form.Field>
              <Form.Input
                label={t('user.edit.username')}
                name='username'
                placeholder={t('user.edit.username_placeholder')}
                onChange={handleInputChange}
                value={username}
                autoComplete='new-password'
              />
            </Form.Field>
            <Form.Field>
              <Form.Input
                label={t('user.edit.password')}
                name='password'
                type={'password'}
                placeholder={t('user.edit.password_placeholder')}
                onChange={handleInputChange}
                value={password}
                autoComplete='new-password'
              />
            </Form.Field>
            <Form.Field>
              <Form.Input
                label={t('user.edit.display_name')}
                name='display_name'
                placeholder={t('user.edit.display_name_placeholder')}
                onChange={handleInputChange}
                value={display_name}
                autoComplete='new-password'
              />
            </Form.Field>
            {userId && (
              <>
                <Form.Field>
                  <Form.Dropdown
                    label={t('user.edit.group')}
                    placeholder={t('user.edit.group_placeholder')}
                    name='group'
                    fluid
                    search
                    selection
                    allowAdditions
                    additionLabel={t('user.edit.group_addition')}
                    onChange={handleInputChange}
                    value={inputs.group}
                    autoComplete='new-password'
                    options={groupOptions}
                  />
                </Form.Field>
                <Form.Field>
                  <Form.Input
                    label={`${t('user.edit.quota')}${renderQuotaWithPrompt(
                      quota,
                      t
                    )}`}
                    name='quota'
                    placeholder={t('user.edit.quota_placeholder')}
                    onChange={handleInputChange}
                    value={quota}
                    type={'number'}
                    autoComplete='new-password'
                  />
                </Form.Field>
              </>
            )}
            <Form.Field>
              <Form.Input
                label={t('user.edit.github_id')}
                name='github_id'
                value={github_id}
                autoComplete='new-password'
                placeholder={t('user.edit.github_id_placeholder')}
                readOnly
              />
            </Form.Field>
            <Form.Field>
              <Form.Input
                label={t('user.edit.wechat_id')}
                name='wechat_id'
                value={wechat_id}
                autoComplete='new-password'
                placeholder={t('user.edit.wechat_id_placeholder')}
                readOnly
              />
            </Form.Field>
            <Form.Field>
              <Form.Input
                label={t('user.edit.email')}
                name='email'
                value={email}
                autoComplete='new-password'
                placeholder={t('user.edit.email_placeholder')}
                readOnly
              />
            </Form.Field>

            {/* TOTP Section */}
            {!userId && (
              <>
                <Divider />
                <Form.Field>
                  <label>Two-Factor Authentication (TOTP)</label>
                  {totpEnabled ? (
                    <Message positive>
                      <Message.Header>TOTP is enabled</Message.Header>
                      <p>Your account is protected with two-factor authentication.</p>
                      <Form.Input
                        placeholder="Enter TOTP code to disable"
                        value={totpCode}
                        onChange={(e) => setTotpCode(e.target.value)}
                        style={{ marginTop: '10px' }}
                      />
                      <Button
                        color="red"
                        onClick={disableTotp}
                        loading={totpLoading}
                        style={{ marginTop: '10px' }}
                      >
                        Disable TOTP
                      </Button>
                    </Message>
                  ) : (
                    <Message info>
                      <Message.Header>TOTP is not enabled</Message.Header>
                      <p>Enable two-factor authentication to secure your account.</p>
                      <Button
                        color="blue"
                        onClick={setupTotp}
                        loading={totpLoading}
                        style={{ marginTop: '10px' }}
                      >
                        Enable TOTP
                      </Button>
                    </Message>
                  )}
                </Form.Field>
              </>
            )}

            {/* Admin TOTP Section - Show when admin is editing other users */}
            {userId && (
              <>
                <Divider />
                <Form.Field>
                  <label>Two-Factor Authentication (TOTP) - Admin Control</label>
                  {totpEnabled ? (
                    <Message warning>
                      <Message.Header>TOTP is enabled for this user</Message.Header>
                      <p>As an administrator, you can disable TOTP for this user if they are locked out.</p>
                      <Button
                        color="red"
                        onClick={adminDisableTotp}
                        loading={totpLoading}
                        style={{ marginTop: '10px' }}
                      >
                        Admin Disable TOTP
                      </Button>
                    </Message>
                  ) : (
                    <Message info>
                      <Message.Header>TOTP is not enabled for this user</Message.Header>
                      <p>This user has not enabled two-factor authentication.</p>
                    </Message>
                  )}
                </Form.Field>
              </>
            )}

            <Button onClick={handleCancel}>
              {t('user.edit.buttons.cancel')}
            </Button>
            <Button positive onClick={submit}>
              {t('user.edit.buttons.submit')}
            </Button>
          </Form>
          )}
        </Card.Content>
      </Card>

      {/* TOTP Setup Modal */}
      <Modal
        open={showTotpSetup}
        onClose={() => setShowTotpSetup(false)}
        size="small"
      >
        <Modal.Header>Setup Two-Factor Authentication</Modal.Header>
        <Modal.Content>
          <Message info>
            <Message.Header>Setup Instructions</Message.Header>
            <ol>
              <li>Install an authenticator app (Google Authenticator, Authy, etc.)</li>
              <li>Scan the QR code below or manually enter the secret key</li>
              <li>Enter the 6-digit code from your authenticator app</li>
              <li>Click "Confirm" to enable TOTP</li>
            </ol>
          </Message>

          {totpQRCode && (
            <div style={{ textAlign: 'center', marginBottom: '20px' }}>
              <Image src={totpQRCode} size="medium" centered />
            </div>
          )}

          <Form.Field>
            <label>Secret Key (manual entry)</label>
            <Form.Input
              value={totpSecret}
              readOnly
              style={{ fontFamily: 'monospace' }}
            />
          </Form.Field>

          <Form.Field>
            <label>Verification Code</label>
            <Form.Input
              placeholder="Enter 6-digit code from your authenticator app"
              value={totpCode}
              onChange={(e) => setTotpCode(e.target.value)}
              maxLength={6}
            />
          </Form.Field>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={() => setShowTotpSetup(false)}>
            Cancel
          </Button>
          <Button
            positive
            onClick={confirmTotp}
            loading={totpLoading}
            disabled={!totpCode || totpCode.length !== 6}
          >
            Confirm
          </Button>
        </Modal.Actions>
      </Modal>
    </div>
  );
};

export default EditUser;
