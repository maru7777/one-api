import React, { useContext, useEffect, useState } from 'react';
import { Card, Grid, Header, Segment } from 'semantic-ui-react';
import { API, showError, showNotice, timestamp2string } from '../../helpers';
import { StatusContext } from '../../context/Status';
import { marked } from 'marked';
import { UserContext } from '../../context/User';
import { Link } from 'react-router-dom';

const Home = () => {
  const [statusState, statusDispatch] = useContext(StatusContext);
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');
  const [userState] = useContext(UserContext);

  const displayNotice = async () => {
    const res = await API.get('/api/notice');
    const { success, message, data } = res.data;
    if (success) {
      let oldNotice = localStorage.getItem('notice');
      if (data !== oldNotice && data !== '') {
        const htmlNotice = marked(data);
        showNotice(htmlNotice, true);
        localStorage.setItem('notice', data);
      }
    } else {
      showError(message);
    }
  };

  const displayHomePageContent = async () => {
    setHomePageContent(localStorage.getItem('home_page_content') || '');
    const res = await API.get('/api/home_page_content');
    const { success, message, data } = res.data;
    if (success) {
      let content = data;
      if (!data.startsWith('https://')) {
        content = marked.parse(data);
      }
      setHomePageContent(content);
      localStorage.setItem('home_page_content', content);
    } else {
      showError(message);
      setHomePageContent('Failed to load homepage content...');
    }
    setHomePageContentLoaded(true);
  };

  const getStartTimeString = () => {
    const timestamp = statusState?.status?.start_time;
    return timestamp2string(timestamp);
  };

  useEffect(() => {
    displayNotice().then();
    displayHomePageContent().then();
  }, []);

  return (
    <>
      {homePageContentLoaded && homePageContent === '' ? (
        <div className='dashboard-container'>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header className='header'>Welcome to One API</Card.Header>
              <Card.Description style={{ lineHeight: '1.6' }}>
                <p>
                  One API is an LLM API interface management and distribution system that helps you better manage and use LLM APIs from various vendors.
                </p>
                {!userState.user && (
                  <p>
                    To use, please <Link to='/login'>log in</Link> or <Link to='/register'>register</Link>.
                  </p>
                )}
              </Card.Description>
            </Card.Content>
          </Card>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header>
                <Header as='h3'>System Status</Header>
              </Card.Header>
              <Grid columns={2} stackable>
                <Grid.Column>
                  <Card
                    fluid
                    className='chart-card'
                    style={{ boxShadow: '0 1px 3px rgba(0,0,0,0.12)' }}
                  >
                    <Card.Content>
                      <Card.Header>
                        <Header as='h3' style={{ color: '#444' }}>
                          System Information
                        </Header>
                      </Card.Header>
                      <Card.Description
                        style={{ lineHeight: '2', marginTop: '1em' }}
                      >
                        <p
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.5em',
                          }}
                        >
                          <i className='info circle icon'></i>
                          <span style={{ fontWeight: 'bold' }}>Name:</span>
                          <span>{statusState?.status?.system_name}</span>
                        </p>
                        <p
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.5em',
                          }}
                        >
                          <i className='code branch icon'></i>
                          <span style={{ fontWeight: 'bold' }}>Version:</span>
                          <span>
                            {statusState?.status?.version || 'unknown'}
                          </span>
                        </p>
                        <p
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.5em',
                          }}
                        >
                          <i className='github icon'></i>
                          <span style={{ fontWeight: 'bold' }}>Source Code:</span>
                          <a
                            href='https://github.com/songquanpeng/one-api'
                            target='_blank'
                            style={{ color: '#2185d0' }}
                          >
                            GitHub Repository
                          </a>
                        </p>
                        <p
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.5em',
                          }}
                        >
                          <i className='clock outline icon'></i>
                          <span style={{ fontWeight: 'bold' }}>Start Time:</span>
                          <span>{getStartTimeString()}</span>
                        </p>
                      </Card.Description>
                    </Card.Content>
                  </Card>
                </Grid.Column>

                <Grid.Column>
                  <Card
                    fluid
                    className='chart-card'
                    style={{ boxShadow: '0 1px 3px rgba(0,0,0,0.12)' }}
                  >
                    <Card.Content>
                      <Card.Header>
                        <Header as='h3' style={{ color: '#444' }}>
                          System Configuration
                        </Header>
                      </Card.Header>
                      <Card.Description
                        style={{ lineHeight: '2', marginTop: '1em' }}
                      >
                        <p
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.5em',
                          }}
                        >
                          <i className='envelope icon'></i>
                          <span style={{ fontWeight: 'bold' }}>Email Verification:</span>
                          <span
                            style={{
                              color: statusState?.status?.email_verification
                                ? '#21ba45'
                                : '#db2828',
                              fontWeight: '500',
                            }}
                          >
                            {statusState?.status?.email_verification
                              ? 'Enabled'
                              : 'Disabled'}
                          </span>
                        </p>
                        <p
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.5em',
                          }}
                        >
                          <i className='github icon'></i>
                          <span style={{ fontWeight: 'bold' }}>
                            GitHub Authentication:
                          </span>
                          <span
                            style={{
                              color: statusState?.status?.github_oauth
                                ? '#21ba45'
                                : '#db2828',
                              fontWeight: '500',
                            }}
                          >
                            {statusState?.status?.github_oauth
                              ? 'Enabled'
                              : 'Disabled'}
                          </span>
                        </p>
                        <p
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.5em',
                          }}
                        >
                          <i className='wechat icon'></i>
                          <span style={{ fontWeight: 'bold' }}>
                            WeChat Authentication:
                          </span>
                          <span
                            style={{
                              color: statusState?.status?.wechat_login
                                ? '#21ba45'
                                : '#db2828',
                              fontWeight: '500',
                            }}
                          >
                            {statusState?.status?.wechat_login
                              ? 'Enabled'
                              : 'Disabled'}
                          </span>
                        </p>
                        <p
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.5em',
                          }}
                        >
                          <i className='shield alternate icon'></i>
                          <span style={{ fontWeight: 'bold' }}>
                            Turnstile Check:
                          </span>
                          <span
                            style={{
                              color: statusState?.status?.turnstile_check
                                ? '#21ba45'
                                : '#db2828',
                              fontWeight: '500',
                            }}
                          >
                            {statusState?.status?.turnstile_check
                              ? 'Enabled'
                              : 'Disabled'}
                          </span>
                        </p>
                      </Card.Description>
                    </Card.Content>
                  </Card>
                </Grid.Column>
              </Grid>
            </Card.Content>
          </Card>{' '}
        </div>
      ) : (
        <>
          {homePageContent.startsWith('https://') ? (
            <iframe
              src={homePageContent}
              style={{ width: '100%', height: '100vh', border: 'none' }}
            />
          ) : (
            <div
              style={{ fontSize: 'larger' }}
              dangerouslySetInnerHTML={{ __html: homePageContent }}
            ></div>
          )}
        </>
      )}
    </>
  );
};

export default Home;
