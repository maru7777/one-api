import PropTypes from 'prop-types';
import { useState, useEffect } from 'react';
import { CHANNEL_OPTIONS } from 'constants/ChannelConstants';
import { useTheme } from '@mui/material/styles';
import { API } from 'utils/api';
import { showError, showSuccess, getChannelModels } from 'utils/common';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Button,
  Divider,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  OutlinedInput,
  ButtonGroup,
  Container,
  Autocomplete,
  FormHelperText,
  Switch,
  Checkbox,
  Box
} from '@mui/material';

import { Formik } from 'formik';
import * as Yup from 'yup';
import { defaultConfig, typeConfig } from '../type/Config'; //typeConfig
import { createFilterOptions } from '@mui/material/Autocomplete';
import CheckBoxOutlineBlankIcon from '@mui/icons-material/CheckBoxOutlineBlank';
import CheckBoxIcon from '@mui/icons-material/CheckBox';

const icon = <CheckBoxOutlineBlankIcon fontSize="small" />;
const checkedIcon = <CheckBoxIcon fontSize="small" />;

const filter = createFilterOptions();
const validationSchema = Yup.object().shape({
  is_edit: Yup.boolean(),
  name: Yup.string().required('名称 不能为空'),
  type: Yup.number().required('渠道 不能为空'),
  key: Yup.string().when(['is_edit', 'type'], {
    is: (is_edit, type) => !is_edit && type !== 33,
    then: Yup.string().required('密钥 不能为空')
  }),
  other: Yup.string(),
  models: Yup.array().min(1, '模型 不能为空'),
  groups: Yup.array().min(1, '用户组 不能为空'),
  base_url: Yup.string().when('type', {
    is: (value) => [3, 8].includes(value),
    then: Yup.string().required('渠道API地址 不能为空'), // base_url 是必需的
    otherwise: Yup.string() // 在其他情况下，base_url 可以是任意字符串
  }),
  model_mapping: Yup.string().test('is-json', '必须是有效的JSON字符串', function (value) {
    try {
      if (value === '' || value === null || value === undefined) {
        return true;
      }
      const parsedValue = JSON.parse(value);
      if (typeof parsedValue === 'object') {
        return true;
      }
    } catch (e) {
      return false;
    }
    return false;
  }),
  model_ratio: Yup.string().nullable().test('is-json', '必须是有效的JSON字符串', function (value) {
    try {
      if (value === '' || value === null || value === undefined) {
        return true;
      }
      const parsedValue = JSON.parse(value);
      if (typeof parsedValue === 'object') {
        return true;
      }
    } catch (e) {
      return false;
    }
    return false;
  }),
  completion_ratio: Yup.string().nullable().test('is-json', '必须是有效的JSON字符串', function (value) {
    try {
      if (value === '' || value === null || value === undefined) {
        return true;
      }
      const parsedValue = JSON.parse(value);
      if (typeof parsedValue === 'object') {
        return true;
      }
    } catch (e) {
      return false;
    }
    return false;
  }),
  inference_profile_arn_map: Yup.string().nullable().test('is-json', '必须是有效的JSON字符串', function (value) {
    try {
      if (value === '' || value === null || value === undefined) {
        return true;
      }
      const parsedValue = JSON.parse(value);
      if (typeof parsedValue === 'object') {
        return true;
      }
    } catch (e) {
      return false;
    }
    return false;
  })
});

const EditModal = ({ open, channelId, onCancel, onOk }) => {
  const theme = useTheme();
  // const [loading, setLoading] = useState(false);
  const [initialInput, setInitialInput] = useState(defaultConfig.input);
  const [inputLabel, setInputLabel] = useState(defaultConfig.inputLabel); //
  const [inputPrompt, setInputPrompt] = useState(defaultConfig.prompt);
  const [groupOptions, setGroupOptions] = useState([]);
  const [modelOptions, setModelOptions] = useState([]);
  const [batchAdd, setBatchAdd] = useState(false);
  const [basicModels, setBasicModels] = useState([]);
  const [defaultPricing, setDefaultPricing] = useState({
    model_ratio: '',
    completion_ratio: '',
  });

  const initChannel = (typeValue) => {
    if (typeConfig[typeValue]?.inputLabel) {
      setInputLabel({
        ...defaultConfig.inputLabel,
        ...typeConfig[typeValue].inputLabel
      });
    } else {
      setInputLabel(defaultConfig.inputLabel);
    }

    if (typeConfig[typeValue]?.prompt) {
      setInputPrompt({
        ...defaultConfig.prompt,
        ...typeConfig[typeValue].prompt
      });
    } else {
      setInputPrompt(defaultConfig.prompt);
    }

    return typeConfig[typeValue]?.input;
  };
  const handleTypeChange = (setFieldValue, typeValue, values) => {
    initChannel(typeValue);
    let localModels = getChannelModels(typeValue);
    setBasicModels(localModels);
    if (localModels.length > 0 && Array.isArray(values['models']) && values['models'].length == 0) {
      setFieldValue('models', initialModel(localModels));
    }

    setFieldValue('config', {});
    // Load default pricing for the new channel type
    loadDefaultPricing(typeValue);
  };

  const fetchGroups = async () => {
    try {
      let res = await API.get(`/api/group/`);
      setGroupOptions(res.data.data);
    } catch (error) {
      showError(error.message);
    }
  };

  const fetchModels = async () => {
    try {
      let res = await API.get(`/api/channel/models`);
      const { data } = res.data;
      data.forEach((item) => {
        if (!item.owned_by) {
          item.owned_by = '未知';
        }
      });
      // 先对data排序
      data.sort((a, b) => {
        const ownedByComparison = a.owned_by.localeCompare(b.owned_by);
        if (ownedByComparison === 0) {
          return a.id.localeCompare(b.id);
        }
        return ownedByComparison;
      });

      setModelOptions(
        data.map((model) => {
          return {
            id: model.id,
            group: model.owned_by
          };
        })
      );
    } catch (error) {
      showError(error.message);
    }
  };

  const loadDefaultPricing = async (channelType) => {
    try {
      const res = await API.get(`/api/channel/default-pricing?type=${channelType}`);
      if (res.data.success) {
        setDefaultPricing({
          model_ratio: res.data.data.model_ratio || '',
          completion_ratio: res.data.data.completion_ratio || '',
        });
      }
    } catch (error) {
      console.error('Failed to load default pricing:', error);
    }
  };

  const submit = async (values, { setErrors, setStatus, setSubmitting }) => {
    setSubmitting(true);
    if (values.base_url && values.base_url.endsWith('/')) {
      values.base_url = values.base_url.slice(0, values.base_url.length - 1);
    }
    if (values.type === 3 && values.other === '') {
      values.other = '2023-09-01-preview';
    }
    if (values.type === 18 && values.other === '') {
      values.other = 'v2.1';
    }
    if (values.key === '') {
      if (values.config.ak && values.config.sk && values.config.region) {
        values.key = `${values.config.ak}|${values.config.sk}|${values.config.region}`;
      } else if (values.config.region && values.config.vertex_ai_project_id && values.config.vertex_ai_adc) {
        values.key = `${values.config.region}|${values.config.vertex_ai_project_id}|${values.config.vertex_ai_adc}`;
      }
    }

    let res;
    const modelsStr = values.models.map((model) => model.id).join(',');
    const configStr = JSON.stringify(values.config);
    values.group = values.groups.join(',');

    // Handle pricing fields - convert empty strings to null for the API
    if (values.model_ratio === '') {
      values.model_ratio = null;
    }
    if (values.completion_ratio === '') {
      values.completion_ratio = null;
    }
    if (channelId) {
      res = await API.put(`/api/channel/`, {
        ...values,
        id: parseInt(channelId),
        models: modelsStr,
        config: configStr
      });
    } else {
      res = await API.post(`/api/channel/`, { ...values, models: modelsStr, config: configStr });
    }
    const { success, message } = res.data;
    if (success) {
      if (channelId) {
        showSuccess('渠道更新成功！');
      } else {
        showSuccess('渠道创建成功！');
      }
      setSubmitting(false);
      setStatus({ success: true });
      onOk(true);
    } else {
      setStatus({ success: false });
      showError(message);
      setErrors({ submit: message });
    }
  };

  function initialModel(channelModel) {
    if (!channelModel) {
      return [];
    }

    // 如果 channelModel 是一个字符串
    if (typeof channelModel === 'string') {
      channelModel = channelModel.split(',');
    }
    let modelList = channelModel.map((model) => {
      const modelOption = modelOptions.find((option) => option.id === model);
      if (modelOption) {
        return modelOption;
      }
      return { id: model, group: '自定义：点击或回车输入' };
    });
    return modelList;
  }

  const loadChannel = async () => {
    let res = await API.get(`/api/channel/${channelId}`);
    const { success, message, data } = res.data;
    if (success) {
      if (data.models === '') {
        data.models = [];
      } else {
        data.models = initialModel(data.models);
      }
      if (data.group === '') {
        data.groups = [];
      } else {
        data.groups = data.group.split(',');
      }
      if (data.model_mapping !== '') {
        data.model_mapping = JSON.stringify(JSON.parse(data.model_mapping), null, 2);
      }
      if (data.config !== '') {
        data.config = JSON.parse(data.config);
      }
      // Format pricing fields for display
      if (data.model_ratio && data.model_ratio !== '') {
        try {
          data.model_ratio = JSON.stringify(JSON.parse(data.model_ratio), null, 2);
        } catch (e) {
          console.error('Failed to parse model_ratio:', e);
        }
      }
      if (data.completion_ratio && data.completion_ratio !== '') {
        try {
          data.completion_ratio = JSON.stringify(JSON.parse(data.completion_ratio), null, 2);
        } catch (e) {
          console.error('Failed to parse completion_ratio:', e);
        }
      }

      data.base_url = data.base_url ?? '';
      data.is_edit = true;
      initChannel(data.type);
      setInitialInput(data);
      // Load default pricing for this channel type
      loadDefaultPricing(data.type);
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    fetchGroups().then();
    fetchModels().then();
  }, []);

  useEffect(() => {
    setBatchAdd(false);
    if (channelId) {
      loadChannel().then();
    } else {
      initChannel(1);
      setInitialInput({ ...defaultConfig.input, is_edit: false });
      // Load default pricing for new channels
      loadDefaultPricing(1);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [channelId]);

  return (
    <Dialog open={open} onClose={onCancel} fullWidth maxWidth={'md'}>
      <DialogTitle
        sx={{
          margin: '0px',
          fontWeight: 700,
          lineHeight: '1.55556',
          padding: '24px',
          fontSize: '1.125rem'
        }}
      >
        {channelId ? '编辑渠道' : '新建渠道'}
      </DialogTitle>
      <Divider />
      <DialogContent>
        <Formik initialValues={initialInput} enableReinitialize validationSchema={validationSchema} onSubmit={submit}>
          {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, touched, values, setFieldValue }) => (
            <form noValidate onSubmit={handleSubmit}>
              <FormControl fullWidth error={Boolean(touched.type && errors.type)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="channel-type-label">{inputLabel.type}</InputLabel>
                <Select
                  id="channel-type-label"
                  label={inputLabel.type}
                  value={values.type}
                  name="type"
                  onBlur={handleBlur}
                  onChange={(e) => {
                    handleChange(e);
                    handleTypeChange(setFieldValue, e.target.value, values);
                  }}
                  MenuProps={{
                    PaperProps: {
                      style: {
                        maxHeight: 200
                      }
                    }
                  }}
                >
                  {Object.values(CHANNEL_OPTIONS)
                    .sort((a, b) => {
                      return a.text.localeCompare(b.text);
                    })
                    .map((option) => {
                      return (
                        <MenuItem key={option.value} value={option.value}>
                          {option.text}
                        </MenuItem>
                      );
                    })}
                </Select>
                {touched.type && errors.type ? (
                  <FormHelperText error id="helper-tex-channel-type-label">
                    {errors.type}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-type-label"> {inputPrompt.type} </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth error={Boolean(touched.name && errors.name)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="channel-name-label">{inputLabel.name}</InputLabel>
                <OutlinedInput
                  id="channel-name-label"
                  label={inputLabel.name}
                  type="text"
                  value={values.name}
                  name="name"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{ autoComplete: 'name' }}
                  aria-describedby="helper-text-channel-name-label"
                />
                {touched.name && errors.name ? (
                  <FormHelperText error id="helper-tex-channel-name-label">
                    {errors.name}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-name-label"> {inputPrompt.name} </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth error={Boolean(touched.base_url && errors.base_url)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="channel-base_url-label">{inputLabel.base_url}</InputLabel>
                <OutlinedInput
                  id="channel-base_url-label"
                  label={inputLabel.base_url}
                  type="text"
                  value={values.base_url}
                  name="base_url"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{}}
                  aria-describedby="helper-text-channel-base_url-label"
                />
                {touched.base_url && errors.base_url ? (
                  <FormHelperText error id="helper-tex-channel-base_url-label">
                    {errors.base_url}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-base_url-label"> {inputPrompt.base_url} </FormHelperText>
                )}
              </FormControl>

              {inputPrompt.other && (
                <FormControl fullWidth error={Boolean(touched.other && errors.other)} sx={{ ...theme.typography.otherInput }}>
                  <InputLabel htmlFor="channel-other-label">{inputLabel.other}</InputLabel>
                  <OutlinedInput
                    id="channel-other-label"
                    label={inputLabel.other}
                    type="text"
                    value={values.other}
                    name="other"
                    onBlur={handleBlur}
                    onChange={handleChange}
                    inputProps={{}}
                    aria-describedby="helper-text-channel-other-label"
                  />
                  {touched.other && errors.other ? (
                    <FormHelperText error id="helper-tex-channel-other-label">
                      {errors.other}
                    </FormHelperText>
                  ) : (
                    <FormHelperText id="helper-tex-channel-other-label"> {inputPrompt.other} </FormHelperText>
                  )}
                </FormControl>
              )}

              <FormControl fullWidth sx={{ ...theme.typography.otherInput }}>
                <Autocomplete
                  multiple
                  id="channel-groups-label"
                  options={groupOptions}
                  value={values.groups}
                  onChange={(e, value) => {
                    const event = {
                      target: {
                        name: 'groups',
                        value: value
                      }
                    };
                    handleChange(event);
                  }}
                  onBlur={handleBlur}
                  filterSelectedOptions
                  renderInput={(params) => <TextField {...params} name="groups" error={Boolean(errors.groups)} label={inputLabel.groups} />}
                  aria-describedby="helper-text-channel-groups-label"
                />
                {errors.groups ? (
                  <FormHelperText error id="helper-tex-channel-groups-label">
                    {errors.groups}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-groups-label"> {inputPrompt.groups} </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth sx={{ ...theme.typography.otherInput }}>
                <Autocomplete
                  multiple
                  freeSolo
                  id="channel-models-label"
                  options={modelOptions}
                  value={values.models}
                  onChange={(e, value) => {
                    const event = {
                      target: {
                        name: 'models',
                        value: value.map((item) => (typeof item === 'string' ? { id: item, group: '自定义：点击或回车输入' } : item))
                      }
                    };
                    handleChange(event);
                  }}
                  onBlur={handleBlur}
                  // filterSelectedOptions
                  disableCloseOnSelect
                  renderInput={(params) => <TextField {...params} name="models" error={Boolean(errors.models)} label={inputLabel.models} />}
                  groupBy={(option) => option.group}
                  getOptionLabel={(option) => {
                    if (typeof option === 'string') {
                      return option;
                    }
                    if (option.inputValue) {
                      return option.inputValue;
                    }
                    return option.id;
                  }}
                  filterOptions={(options, params) => {
                    const filtered = filter(options, params);
                    const { inputValue } = params;
                    const isExisting = options.some((option) => inputValue === option.id);
                    if (inputValue !== '' && !isExisting) {
                      filtered.push({
                        id: inputValue,
                        group: '自定义：点击或回车输入'
                      });
                    }
                    return filtered;
                  }}
                  renderOption={(props, option, { selected }) => (
                    <li {...props}>
                      <Checkbox icon={icon} checkedIcon={checkedIcon} style={{ marginRight: 8 }} checked={selected} />
                      {option.id}
                    </li>
                  )}
                />
                {errors.models ? (
                  <FormHelperText error id="helper-tex-channel-models-label">
                    {errors.models}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-models-label"> {inputPrompt.models} </FormHelperText>
                )}
              </FormControl>
              <Container
                sx={{
                  textAlign: 'right'
                }}
              >
                <ButtonGroup variant="outlined" aria-label="small outlined primary button group">
                  <Button
                    onClick={() => {
                      setFieldValue('models', initialModel(basicModels));
                    }}
                  >
                    填入相关模型
                  </Button>
                  <Button
                    onClick={() => {
                      setFieldValue('models', modelOptions);
                    }}
                  >
                    填入所有模型
                  </Button>
                </ButtonGroup>
              </Container>
              {inputLabel.key && (
                <>
                  <FormControl fullWidth error={Boolean(touched.key && errors.key)} sx={{ ...theme.typography.otherInput }}>
                    {!batchAdd ? (
                      <>
                        <InputLabel htmlFor="channel-key-label">{inputLabel.key}</InputLabel>
                        <OutlinedInput
                          id="channel-key-label"
                          label={inputLabel.key}
                          type="text"
                          value={values.key}
                          name="key"
                          onBlur={handleBlur}
                          onChange={handleChange}
                          inputProps={{}}
                          aria-describedby="helper-text-channel-key-label"
                        />
                      </>
                    ) : (
                      <TextField
                        multiline
                        id="channel-key-label"
                        label={inputLabel.key}
                        value={values.key}
                        name="key"
                        onBlur={handleBlur}
                        onChange={handleChange}
                        aria-describedby="helper-text-channel-key-label"
                        minRows={5}
                        placeholder={inputPrompt.key + '，一行一个密钥'}
                      />
                    )}

                    {touched.key && errors.key ? (
                      <FormHelperText error id="helper-tex-channel-key-label">
                        {errors.key}
                      </FormHelperText>
                    ) : (
                      <FormHelperText id="helper-tex-channel-key-label"> {inputPrompt.key} </FormHelperText>
                    )}
                  </FormControl>
                  {channelId === 0 && (
                    <Container
                      sx={{
                        textAlign: 'right'
                      }}
                    >
                      <Switch checked={batchAdd} onChange={(e) => setBatchAdd(e.target.checked)} />
                      批量添加
                    </Container>
                  )}
                </>
              )}

              {inputLabel.config &&
                Object.keys(inputLabel.config).map((configName) => {
                  return (
                    <FormControl key={'config.' + configName} fullWidth sx={{ ...theme.typography.otherInput }}>
                      <TextField
                        multiline
                        key={'config.' + configName}
                        name={'config.' + configName}
                        value={values.config?.[configName] || ''}
                        label={configName}
                        placeholder={inputPrompt.config[configName]}
                        onChange={handleChange}
                      />
                      <FormHelperText id={`helper-tex-config.${configName}-label`}> {inputPrompt.config[configName]} </FormHelperText>
                    </FormControl>
                  );
                })}

              <FormControl fullWidth error={Boolean(touched.model_mapping && errors.model_mapping)} sx={{ ...theme.typography.otherInput }}>
                {/* <InputLabel htmlFor="channel-model_mapping-label">{inputLabel.model_mapping}</InputLabel> */}
                <TextField
                  multiline
                  id="channel-model_mapping-label"
                  label={inputLabel.model_mapping}
                  value={values.model_mapping}
                  name="model_mapping"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  aria-describedby="helper-text-channel-model_mapping-label"
                  minRows={5}
                  placeholder={inputPrompt.model_mapping}
                />
                {touched.model_mapping && errors.model_mapping ? (
                  <FormHelperText error id="helper-tex-channel-model_mapping-label">
                    {errors.model_mapping}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-model_mapping-label"> {inputPrompt.model_mapping} </FormHelperText>
                )}
              </FormControl>
              <FormControl fullWidth error={Boolean(touched.system_prompt && errors.system_prompt)} sx={{ ...theme.typography.otherInput }}>
                {/* <InputLabel htmlFor="channel-model_mapping-label">{inputLabel.model_mapping}</InputLabel> */}
                <TextField
                  multiline
                  id="channel-system_prompt-label"
                  label={inputLabel.system_prompt}
                  value={values.system_prompt}
                  name="system_prompt"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  aria-describedby="helper-text-channel-system_prompt-label"
                  minRows={5}
                  placeholder={inputPrompt.system_prompt}
                />
                {touched.system_prompt && errors.system_prompt ? (
                  <FormHelperText error id="helper-tex-channel-system_prompt-label">
                    {errors.system_prompt}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-system_prompt-label"> {inputPrompt.system_prompt} </FormHelperText>
                )}
              </FormControl>

              {/* Channel-specific pricing fields */}
              <FormControl fullWidth error={Boolean(touched.model_ratio && errors.model_ratio)} sx={{ ...theme.typography.otherInput }}>
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
                  <InputLabel
                    htmlFor="channel-model_ratio-label"
                    sx={{
                      position: 'relative',
                      transform: 'none',
                      fontSize: '0.875rem',
                      fontWeight: 500,
                      color: theme.palette.text.primary
                    }}
                  >
                    {inputLabel.model_ratio}
                  </InputLabel>
                  <Button
                    size="small"
                    variant="outlined"
                    onClick={() => {
                      // Format the JSON string for better display
                      let formattedValue = defaultPricing.model_ratio;
                      if (formattedValue && formattedValue !== '') {
                        try {
                          const parsed = JSON.parse(formattedValue);
                          formattedValue = JSON.stringify(parsed, null, 2);
                        } catch (e) {
                          console.error('Failed to format model_ratio JSON:', e);
                        }
                      }

                      setFieldValue('model_ratio', formattedValue);
                    }}
                  >
                    加载默认值
                  </Button>
                </Box>
                <TextField
                  multiline
                  id="channel-model_ratio-label"
                  value={values.model_ratio}
                  name="model_ratio"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  aria-describedby="helper-text-channel-model_ratio-label"
                  minRows={5}
                  placeholder={inputPrompt.model_ratio}
                />
                {touched.model_ratio && errors.model_ratio ? (
                  <FormHelperText error id="helper-tex-channel-model_ratio-label">
                    {errors.model_ratio}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-model_ratio-label">
                    JSON 格式：{`{"模型名称": 价格倍率}`}。价格倍率乘以 token 数量计算费用。
                  </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth error={Boolean(touched.completion_ratio && errors.completion_ratio)} sx={{ ...theme.typography.otherInput }}>
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
                  <InputLabel
                    htmlFor="channel-completion_ratio-label"
                    sx={{
                      position: 'relative',
                      transform: 'none',
                      fontSize: '0.875rem',
                      fontWeight: 500,
                      color: theme.palette.text.primary
                    }}
                  >
                    {inputLabel.completion_ratio}
                  </InputLabel>
                  <Button
                    size="small"
                    variant="outlined"
                    onClick={() => {
                      // Format the JSON string for better display
                      let formattedValue = defaultPricing.completion_ratio;
                      if (formattedValue && formattedValue !== '') {
                        try {
                          const parsed = JSON.parse(formattedValue);
                          formattedValue = JSON.stringify(parsed, null, 2);
                        } catch (e) {
                          console.error('Failed to format completion_ratio JSON:', e);
                        }
                      }

                      setFieldValue('completion_ratio', formattedValue);
                    }}
                  >
                    加载默认值
                  </Button>
                </Box>
                <TextField
                  multiline
                  id="channel-completion_ratio-label"
                  value={values.completion_ratio}
                  name="completion_ratio"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  aria-describedby="helper-text-channel-completion_ratio-label"
                  minRows={5}
                  placeholder={inputPrompt.completion_ratio}
                />
                {touched.completion_ratio && errors.completion_ratio ? (
                  <FormHelperText error id="helper-tex-channel-completion_ratio-label">
                    {errors.completion_ratio}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-completion_ratio-label">
                    JSON 格式：{`{"模型名称": 输出倍率}`}。输出倍率乘以输出 token 数量。
                  </FormHelperText>
                )}
              </FormControl>

              {/* AWS-specific inference profile ARN mapping */}
              {values.type === 33 && (
                <FormControl fullWidth error={Boolean(touched.inference_profile_arn_map && errors.inference_profile_arn_map)} sx={{ ...theme.typography.otherInput }}>
                  <InputLabel
                    htmlFor="channel-inference_profile_arn_map-label"
                    sx={{
                      position: 'relative',
                      transform: 'none',
                      fontSize: '0.875rem',
                      fontWeight: 500,
                      color: theme.palette.text.primary
                    }}
                  >
                    {inputLabel.inference_profile_arn_map}
                  </InputLabel>
                  <TextField
                    multiline
                    id="channel-inference_profile_arn_map-label"
                    value={values.inference_profile_arn_map}
                    name="inference_profile_arn_map"
                    onBlur={handleBlur}
                    onChange={handleChange}
                    aria-describedby="helper-text-channel-inference_profile_arn_map-label"
                    minRows={5}
                    placeholder={inputPrompt.inference_profile_arn_map}
                  />
                  {touched.inference_profile_arn_map && errors.inference_profile_arn_map ? (
                    <FormHelperText error id="helper-tex-channel-inference_profile_arn_map-label">
                      {errors.inference_profile_arn_map}
                    </FormHelperText>
                  ) : (
                    <FormHelperText id="helper-tex-channel-inference_profile_arn_map-label">
                      JSON 格式：{`{"模型名称": "arn:aws:bedrock:region:account:inference-profile/profile-id"}`}。将模型名称映射到 AWS Bedrock 推理配置文件 ARN。留空则使用默认模型 ID。
                    </FormHelperText>
                  )}
                </FormControl>
              )}

              <DialogActions>
                <Button onClick={onCancel}>取消</Button>
                <Button disableElevation disabled={isSubmitting} type="submit" variant="contained" color="primary">
                  提交
                </Button>
              </DialogActions>
            </form>
          )}
        </Formik>
      </DialogContent>
    </Dialog>
  );
};

export default EditModal;

EditModal.propTypes = {
  open: PropTypes.bool,
  channelId: PropTypes.number,
  onCancel: PropTypes.func,
  onOk: PropTypes.func
};
