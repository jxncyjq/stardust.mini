package errors

// Try 模拟 try/catch/finally 结构
func TryFunc(fun func(), catch func(err interface{}), finally func()) {
	defer func() {
		// 先执行 catch，再执行 finally（符合标准语义）
		if r := recover(); r != nil {
			if catch != nil {
				catch(r)
			} else {
				panic(r)
			}
		}

		// finally 块无论如何都会执行
		if finally != nil {
			finally()
		}
	}()

	// 执行 try 块
	fun()
}
