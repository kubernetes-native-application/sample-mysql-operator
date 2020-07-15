package controllers

import (
	"context"
	mysqlv1alpha1 "sample-mysql-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const MYSQL_FINALIZER = "finalizer.sample-mysql-operator.com"

func (r *MySQLReconciler) finalize(mysql *mysqlv1alpha1.MySQL) error {
	// MySQL이 삭제되었는지 확인
	isMysqlMarkedToBeDeleted := mysql.GetDeletionTimestamp() != nil
	if isMysqlMarkedToBeDeleted {
		// 파이널라이저가 남아있는 경우
		if contains(mysql.GetFinalizers(), MYSQL_FINALIZER) {
			// 클린업 로직 수행
			if err := cleanUp(r, mysql); err != nil {
				return err
			}
			// 파이널라이저를 제거
			controllerutil.RemoveFinalizer(mysql, MYSQL_FINALIZER)
			return r.Update(context.TODO(), mysql)
		}
	}
	// 파이널라이저가 없는 경우 추가
	if !contains(mysql.GetFinalizers(), MYSQL_FINALIZER) {
		controllerutil.AddFinalizer(mysql, MYSQL_FINALIZER)
		return r.Update(context.TODO(), mysql)
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func cleanUp(r *MySQLReconciler, mysql *mysqlv1alpha1.MySQL) error {
	// 추후 MySQL을 제거할 때 필요한 로직이 있으면 여기에 추가한다
	r.Log.Info("Successfully clean up")
	return nil
}
