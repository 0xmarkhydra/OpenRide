import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flashx_rider/main.dart';

void main() {
  setUp(() {
    FlutterSecureStorage.setMockInitialValues({});
  });

  testWidgets('Rider app boots to phone authentication when no session exists',
      (tester) async {
    await tester.pumpWidget(const FlashXRiderApp());
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));
    expect(find.text('Đi đâu cũng nhanh hơn.'), findsOneWidget);
    expect(find.text('Số điện thoại'), findsOneWidget);
    expect(find.text('Tiếp tục'), findsOneWidget);
  });
}
