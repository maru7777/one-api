import React, { useEffect, useState } from 'react';
import { Button, Form, Header, Segment } from 'semantic-ui-react';
import { useParams, useNavigate } from 'react-router-dom';
import { API, showError, showSuccess } from '../../helpers';
import { renderQuota, renderQuotaWithPrompt } from '../../helpers/render';

const EditUser = () => {
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
    group: 'default'
  });
  const [groupOptions, setGroupOptions] = useState([]);
  const { username, display_name, password, github_id, wechat_id, email, quota, group } =
    inputs;
  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };
  const fetchGroups = async () => {
    try {
      let res = await API.get(`/api/group/`);
      setGroupOptions(res.data.data.map((group) => ({
        key: group,
        text: group,
        value: group,
      })));
    } catch (error) {
      showError(error.message);
    }
  };
  const navigate = useNavigate();
  const handleCancel = () => {
    navigate("/setting");
  }
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
    } else {
      showError(message);
    }
    setLoading(false);
  };
  useEffect(() => {
    loadUser().then();
    if (userId) {
      fetchGroups().then();
    }
  }, []);

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
      showSuccess('User information updated successfully!');
    } else {
      showError(message);
    }
  };

  return (
    <>
      <Segment loading={loading}>
        <Header as='h3'>Update user information</Header>
        <Form autoComplete='new-password'>
          <Form.Field>
            <Form.Input
              label='Username'
              name='username'
              placeholder={'Please enter a new username'}
              onChange={handleInputChange}
              value={username}
              autoComplete='new-password'
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label='Password'
              name='password'
              type={'password'}
              placeholder={'Please enter a new password, at least 8 characters'}
              onChange={handleInputChange}
              value={password}
              autoComplete='new-password'
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label='Display name'
              name='display_name'
              placeholder={'Please enter a new display name'}
              onChange={handleInputChange}
              value={display_name}
              autoComplete='new-password'
            />
          </Form.Field>
          {
            userId && <>
              <Form.Field>
                <Form.Dropdown
                  label='Group'
                  placeholder={'Please select a group'}
                  name='group'
                  fluid
                  search
                  selection
                  allowAdditions
                  additionLabel={'Please edit the group rate on the system settings page to add a new group:'}
                  onChange={handleInputChange}
                  value={inputs.group}
                  autoComplete='new-password'
                  options={groupOptions}
                />
              </Form.Field>
              <Form.Field>
                <Form.Input
                  label={`Remaining quota${renderQuotaWithPrompt(quota)}`}
                  name='quota'
                  placeholder={'Please enter a new remaining quota'}
                  onChange={handleInputChange}
                  value={quota}
                  type={'number'}
                  autoComplete='new-password'
                />
              </Form.Field>
            </>
          }
          <Form.Field>
            <Form.Input
              label='Bound GitHub account'
              name='github_id'
              value={github_id}
              autoComplete='new-password'
              placeholder='This item is read-only, users need to bind through the relevant binding button on the personal settings page, cannot be directly modified'
              readOnly
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label='Bound WeChat account'
              name='wechat_id'
              value={wechat_id}
              autoComplete='new-password'
              placeholder='This item is read-only, users need to bind through the relevant binding button on the personal settings page, cannot be directly modified'
              readOnly
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label='Bound email account'
              name='email'
              value={email}
              autoComplete='new-password'
              placeholder='This item is read-only, users need to bind through the relevant binding button on the personal settings page, cannot be directly modified'
              readOnly
            />
          </Form.Field>
          <Button onClick={handleCancel}>Cancel</Button>
          <Button positive onClick={submit}>Submit</Button>
        </Form>
      </Segment>
    </>
  );
};

export default EditUser;
