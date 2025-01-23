import React, { useEffect, useState } from 'react';
import { Button, Dropdown, Form, Input, Label, Message, Pagination, Popup, Table } from 'semantic-ui-react';
import { Link } from 'react-router-dom';
import {
  API,
  loadChannelModels,
  setPromptShown,
  shouldShowPrompt,
  showError,
  showInfo,
  showSuccess,
  timestamp2string
} from '../helpers';

import { CHANNEL_OPTIONS, ITEMS_PER_PAGE } from '../constants';
import { renderGroup, renderNumber } from '../helpers/render';

function renderTimestamp(timestamp) {
  return (
    <>
      {timestamp2string(timestamp)}
    </>
  );
}

let type2label = undefined;

function renderType(type) {
  if (!type2label) {
    type2label = new Map;
    for (let i = 0; i < CHANNEL_OPTIONS.length; i++) {
      type2label[CHANNEL_OPTIONS[i].value] = CHANNEL_OPTIONS[i];
    }
    type2label[0] = { value: 0, text: 'Unknown type', color: 'grey' };
  }
  return <Label basic color={type2label[type]?.color}>{type2label[type] ? type2label[type].text : type}</Label>;
}

function renderBalance(type, balance) {
  switch (type) {
    case 1: // OpenAI
      return <span>${balance.toFixed(2)}</span>;
    case 4: // CloseAI
      return <span>¥{balance.toFixed(2)}</span>;
    case 8: // Custom
      return <span>${balance.toFixed(2)}</span>;
    case 5: // OpenAI-SB
      return <span>¥{(balance / 10000).toFixed(2)}</span>;
    case 10: // AI Proxy
      return <span>{renderNumber(balance)}</span>;
    case 12: // API2GPT
      return <span>¥{balance.toFixed(2)}</span>;
    case 13: // AIGC2D
      return <span>{renderNumber(balance)}</span>;
    case 36: // DeepSeek
      return <span>¥{balance.toFixed(2)}</span>;
    case 44: // SiliconFlow
      return <span>¥{balance.toFixed(2)}</span>;
    default:
      return <span>Not supported</span>;
  }
}

function isShowDetail() {
  return localStorage.getItem("show_detail") === "true";
}

const promptID = "detail"

const ChannelsTable = () => {
  const [channels, setChannels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [updatingBalance, setUpdatingBalance] = useState(false);
  const [showPrompt, setShowPrompt] = useState(shouldShowPrompt(promptID));
  const [showDetail, setShowDetail] = useState(isShowDetail());

  const loadChannels = async (startIdx) => {
    const res = await API.get(`/api/channel/?p=${startIdx}`);
    const { success, message, data } = res.data;
    if (success) {
        let localChannels = data.map((channel) => {
            if (channel.models === '') {
                channel.models = [];
                channel.test_model = "";
            } else {
                channel.models = channel.models.split(',');
                if (channel.models.length > 0) {
                    channel.test_model = channel.models[0];
                }
                channel.model_options = channel.models.map((model) => {
                    return {
                        key: model,
                        text: model,
                        value: model,
                    }
                })
                console.log('channel', channel)
            }
            return channel;
        });
        if (startIdx === 0) {
            setChannels(localChannels);
        } else {
            let newChannels = [...channels];
            newChannels.splice(startIdx * ITEMS_PER_PAGE, data.length, ...localChannels);
            setChannels(newChannels);
        }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(channels.length / ITEMS_PER_PAGE) + 1) {
        // In this case we have to load more data and then append them.
        await loadChannels(activePage - 1);
      }
      setActivePage(activePage);
    })();
  };

  const refresh = async () => {
    setLoading(true);
    await loadChannels(activePage - 1);
  };

  const toggleShowDetail = () => {
    setShowDetail(!showDetail);
    localStorage.setItem("show_detail", (!showDetail).toString());
  }

  useEffect(() => {
    loadChannels(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
    loadChannelModels().then();
  }, []);

  const manageChannel = async (id, action, idx, value) => {
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/channel/${id}/`);
        break;
      case 'enable':
        data.status = 1;
        res = await API.put('/api/channel/', data);
        break;
      case 'disable':
        data.status = 2;
        res = await API.put('/api/channel/', data);
        break;
      case 'priority':
        if (value === '') {
          return;
        }
        data.priority = parseInt(value);
        res = await API.put('/api/channel/', data);
        break;
      case 'weight':
        if (value === '') {
          return;
        }
        data.weight = parseInt(value);
        if (data.weight < 0) {
          data.weight = 0;
        }
        res = await API.put('/api/channel/', data);
        break;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess('Operation successfully completed!');
      let channel = res.data.data;
      let newChannels = [...channels];
      let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      if (action === 'delete') {
        newChannels[realIdx].deleted = true;
      } else {
        newChannels[realIdx].status = channel.status;
      }
      setChannels(newChannels);
    } else {
      showError(message);
    }
  };

  const renderStatus = (status) => {
    switch (status) {
      case 1:
        return <Label basic color='green'>Enabled</Label>;
      case 2:
        return (
          <Popup
            trigger={<Label basic color='red'>
              Disabled
            </Label>}
            content='本Channel被手动Disable'
            basic
          />
        );
      case 3:
        return (
          <Popup
            trigger={<Label basic color='yellow'>
              Disabled
            </Label>}
            content='本Channel被程序自动Disable'
            basic
          />
        );
      default:
        return (
          <Label basic color='grey'>
            Unknown status
          </Label>
        );
    }
  };

  const renderResponseTime = (responseTime) => {
    let time = responseTime / 1000;
    time = time.toFixed(2) + 's';
    if (responseTime === 0) {
      return <Label basic color='grey'>Not tested</Label>;
    } else if (responseTime <= 1000) {
      return <Label basic color='green'>{time}</Label>;
    } else if (responseTime <= 3000) {
      return <Label basic color='olive'>{time}</Label>;
    } else if (responseTime <= 5000) {
      return <Label basic color='yellow'>{time}</Label>;
    } else {
      return <Label basic color='red'>{time}</Label>;
    }
  };

  const searchChannels = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load files instead.
      await loadChannels(0);
      setActivePage(1);
      return;
    }
    setSearching(true);
    const res = await API.get(`/api/channel/search?keyword=${searchKeyword}`);
    const { success, message, data } = res.data;
    if (success) {
      setChannels(data);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const switchTestModel = async (idx, model) => {
    let newChannels = [...channels];
    let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
    newChannels[realIdx].test_model = model;
    setChannels(newChannels);
  };

  const testChannel = async (id, name, idx, m) => {
    const res = await API.get(`/api/channel/test/${id}?model=${m}`);
    const { success, message, time, model } = res.data;
    if (success) {
      let newChannels = [...channels];
      let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      newChannels[realIdx].response_time = time * 1000;
      newChannels[realIdx].test_time = Date.now() / 1000;
      setChannels(newChannels);
      showInfo(`Channel ${name} Test成功，Model ${model}，耗时 ${time.toFixed(2)}s。`);
    } else {
      showError(message);
    }
    let newChannels = [...channels];
    let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
    newChannels[realIdx].response_time = time * 1000;
    newChannels[realIdx].test_time = Date.now() / 1000;
    setChannels(newChannels);
  };

  const testChannels = async (scope) => {
    const res = await API.get(`/api/channel/test?scope=${scope}`);
    const { success, message } = res.data;
    if (success) {
      showInfo('已成功开始TestChannel，请Refresh页面查看结果。');
    } else {
      showError(message);
    }
  };

  const deleteAllDisabledChannels = async () => {
    const res = await API.delete(`/api/channel/disabled`);
    const { success, message, data } = res.data;
    if (success) {
      showSuccess(`已Delete所有DisableChannel，共计 ${data} 个`);
      await refresh();
    } else {
      showError(message);
    }
  };

  const updateChannelBalance = async (id, name, idx) => {
    const res = await API.get(`/api/channel/update_balance/${id}/`);
    const { success, message, balance } = res.data;
    if (success) {
      let newChannels = [...channels];
      let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      newChannels[realIdx].balance = balance;
      newChannels[realIdx].balance_updated_time = Date.now() / 1000;
      setChannels(newChannels);
      showInfo(`Channel ${name} balance updated successfully!`);
    } else {
      showError(message);
    }
  };

  const updateAllChannelsBalance = async () => {
    setUpdatingBalance(true);
    const res = await API.get(`/api/channel/update_balance`);
    const { success, message } = res.data;
    if (success) {
      showInfo('The balance of all enabled channels has been updated!');
    } else {
      showError(message);
    }
    setUpdatingBalance(false);
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortChannel = (key) => {
    if (channels.length === 0) return;
    setLoading(true);
    let sortedChannels = [...channels];
    sortedChannels.sort((a, b) => {
      if (!isNaN(a[key])) {
        // If the value is numeric, subtract to sort
        return a[key] - b[key];
      } else {
        // If the value is not numeric, sort as strings
        return ('' + a[key]).localeCompare(b[key]);
      }
    });
    if (sortedChannels[0].id === channels[0].id) {
      sortedChannels.reverse();
    }
    setChannels(sortedChannels);
    setLoading(false);
  };


  return (
    <>
      <Form onSubmit={searchChannels}>
        <Form.Input
          icon='search'
          fluid
          iconPosition='left'
          placeholder='Search for channel ID, name and key ...'
          value={searchKeyword}
          loading={searching}
          onChange={handleKeywordChange}
        />
      </Form>
      {
        showPrompt && (
          <Message onDismiss={() => {
            setShowPrompt(false);
            setPromptShown(promptID);
          }}>
            OpenAI Channel no longer supports getting Balance via key, so Balance is shown as 0. For supported ChannelTypes, please click Balance to Refresh.
            <br/>
            ChannelTest only supports chat Models, preferring gpt-3.5-turbo. If this Model is not available, it will use the first Model in your configured Model list.
            <br/>
            Click the Details button below to display Balance and additional TestModel Settings.
          </Message>
      )
    }
      <Table basic compact size='small'>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('id');
              }}
            >
              ID
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('name');
              }}
            >
              Name
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('group');
              }}
            >
              Group
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('type');
              }}
            >
              Type
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('status');
              }}
            >
              Status
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('response_time');
              }}
            >
              Response time
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('balance');
              }}
              hidden={!showDetail}
            >
              Balance
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('priority');
              }}
            >
              Priority
            </Table.HeaderCell>
            <Table.HeaderCell hidden={!showDetail}>TestModel</Table.HeaderCell>
            <Table.HeaderCell>Operation</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {channels
            .slice(
              (activePage - 1) * ITEMS_PER_PAGE,
              activePage * ITEMS_PER_PAGE
            )
            .map((channel, idx) => {
              if (channel.deleted) return <></>;
              return (
                <Table.Row key={channel.id}>
                  <Table.Cell>{channel.id}</Table.Cell>
                  <Table.Cell>{channel.name ? channel.name : 'None'}</Table.Cell>
                  <Table.Cell>{renderGroup(channel.group)}</Table.Cell>
                  <Table.Cell>{renderType(channel.type)}</Table.Cell>
                  <Table.Cell>{renderStatus(channel.status)}</Table.Cell>
                  <Table.Cell>
                    <Popup
                      content={channel.test_time ? renderTimestamp(channel.test_time) : 'Not tested'}
                      key={channel.id}
                      trigger={renderResponseTime(channel.response_time)}
                      basic
                    />
                  </Table.Cell>
                  <Table.Cell hidden={!showDetail}>
                    <Popup
                      trigger={<span onClick={() => {
                        updateChannelBalance(channel.id, channel.name, idx);
                      }} style={{ cursor: 'pointer' }}>
                      {renderBalance(channel.type, channel.balance)}
                    </span>}
                      content='Click to refresh'
                      basic
                    />
                  </Table.Cell>
                  <Table.Cell>
                    <Popup
                      trigger={<Input type='number' defaultValue={channel.priority} onBlur={(event) => {
                        manageChannel(
                          channel.id,
                          'priority',
                          idx,
                          event.target.value
                        );
                      }}>
                        <input style={{ maxWidth: '60px' }} />
                      </Input>}
                      content='Channel priority - higher value means higher priority'
                    />
                  </Table.Cell>
                  <Table.Cell hidden={!showDetail}>
                    <Dropdown
                      placeholder='请选择TestModel'
                      selection
                      options={channel.model_options}
                      defaultValue={channel.test_model}
                      onChange={(event, data) => {
                        switchTestModel(idx, data.value);
                      }}
                    />
                  </Table.Cell>
                  <Table.Cell>
                    <div>
                      <Button
                        size={'small'}
                        positive
                        onClick={() => {
                          testChannel(channel.id, channel.name, idx, channel.test_model);
                        }}
                      >
                        Test
                      </Button>
                      {/*<Button*/}
                      {/*  size={'small'}*/}
                      {/*  positive*/}
                      {/*  loading={updatingBalance}*/}
                      {/*  onClick={() => {*/}
                      {/*    updateChannelBalance(channel.id, channel.name, idx);*/}
                      {/*  }}*/}
                      {/*>*/}
                      {/*  Update balance*/}
                      {/*</Button>*/}
                      <Popup
                        trigger={
                          <Button size='small' negative>
                            Delete
                          </Button>
                        }
                        on='click'
                        flowing
                        hoverable
                      >
                        <Button
                          negative
                          onClick={() => {
                            manageChannel(channel.id, 'delete', idx);
                          }}
                        >
                          Delete channel {channel.name}
                        </Button>
                      </Popup>
                      <Button
                        size={'small'}
                        onClick={() => {
                          manageChannel(
                            channel.id,
                            channel.status === 1 ? 'disable' : 'enable',
                            idx
                          );
                        }}
                      >
                        {channel.status === 1 ? 'Disable' : 'Enable'}
                      </Button>
                      <Button
                        size={'small'}
                        as={Link}
                        to={'/channel/edit/' + channel.id}
                      >
                        Edit
                      </Button>
                    </div>
                  </Table.Cell>
                </Table.Row>
              );
            })}
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan={showDetail ? "10" : "8"}>
              <Button size='small' as={Link} to='/channel/add' loading={loading}>
                Add a new channel
              </Button>
              <Button size='small' loading={loading} onClick={()=>{testChannels("all")}}>
                Test all channels
              </Button>
              <Button size='small' loading={loading} onClick={()=>{testChannels("disabled")}}>
                TestDisableChannel
              </Button>
              {/*<Button size='small' onClick={updateAllChannelsBalance}*/}
              {/*        loading={loading || updatingBalance}>Update the balance of enabled channels</Button>*/}
              <Popup
                trigger={
                  <Button size='small' loading={loading}>
                    DeleteDisableChannel
                  </Button>
                }
                on='click'
                flowing
                hoverable
              >
                <Button size='small' loading={loading} negative onClick={deleteAllDisabledChannels}>
                  Confirm deletion
                </Button>
              </Popup>
              <Pagination
                floated='right'
                activePage={activePage}
                onPageChange={onPaginationChange}
                size='small'
                siblingRange={1}
                totalPages={
                  Math.ceil(channels.length / ITEMS_PER_PAGE) +
                  (channels.length % ITEMS_PER_PAGE === 0 ? 1 : 0)
                }
              />
              <Button size='small' onClick={refresh} loading={loading}>Refresh</Button>
              <Button size='small' onClick={toggleShowDetail}>{showDetail ? "隐藏Details" : "Details"}</Button>
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
    </>
  );
};

export default ChannelsTable;
