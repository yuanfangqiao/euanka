package internal

import (
	"os"

	"go.uber.org/zap/zapcore"
)

var FileRotateLogs = new(fileRotateLogs)

type fileRotateLogs struct{}

// GetWriteSyncer 获取 zapcore.WriteSyncer
func (r *fileRotateLogs) GetWriteSyncer(level string) (zapcore.WriteSyncer, error) {
	// fileWriter, err := rotatelogs.New(
	// 	path.Join(global.CONFIG.Zap.Director, "%Y-%m-%d", level+".log"),
	// 	rotatelogs.WithClock(rotatelogs.Local),
	// 	rotatelogs.WithMaxAge(time.Duration(global.CONFIG.Zap.MaxAge)*24*time.Hour), // 日志留存时间
	// 	rotatelogs.WithRotationTime(time.Hour*24),
	// )
	// if global.CONFIG.Zap.LogInConsole {
	// 	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
	// }
	// return zapcore.AddSync(fileWriter), err

	// 不输出到文件
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), nil
}
