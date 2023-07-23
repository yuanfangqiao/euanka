
## 介绍
这是一个即将起步的AI对话服务体系，当然当前都是构思完善中....

### 依赖项目

#### 客户端
树莓派 语音终端（GO实现） https://github.com/yuanfangqiao/euanka-client.git

#### ASR
websocket 流式识别

sherpa-onnx-euanka https://github.com/yuanfangqiao/sherpa-onnx-euanka

sherpa-onnx https://github.com/k2-fsa/sherpa-onnx

docker镜像：

registry.cn-shenzhen.aliyuncs.com/yuanfangqiao/k2-sherpa-onnx:0.1

#### NLP
传统NLP，量化，意图，词槽，对话

Rasa

euanka-rasa https://github.com/yuanfangqiao/euanka-rasa

rasa https://github.com/RasaHQ/rasa

spacy https://github.com/explosion/spaCy

spacy-models https://github.com/explosion/spacy-models

#### LLM 
大模型，生成式对话AI

chatglm.cpp(yuanfangqiao) https://github.com/yuanfangqiao/chatglm.cpp

chatglm.cpp https://github.com/li-plus/chatglm.cpp

docker镜像：
携带量化加速模型，CPU运行，如果要GPU运行，请参照对应的重新部署

registry.cn-shenzhen.aliyuncs.com/yuanfangqiao/chatglm2-6b-cpp-py-all:0.1 

#### TTS
vits 文字转语音，声音克隆

VITS-Umamusume-voice-synthesizer(yuanfangqiao) https://github.com/yuanfangqiao/VITS-Umamusume-voice-synthesizer

VITS-Umamusume-voice-synthesizer https://huggingface.co/spaces/Plachta/VITS-Umamusume-voice-synthesizer

！！！ containerd镜像：就是huggingface原生镜像，采用containerd启动,CPU运行

registry.hf.space/plachta-vits-umamusume-voice-synthesizer:latest

### 构建

```
go build src/main.go
```

### 运行
```shell
./main
```
