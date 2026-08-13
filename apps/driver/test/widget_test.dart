import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flashx_driver/main.dart';

void main() {
  setUp(() {
    FlutterSecureStorage.setMockInitialValues({});
  });

  testWidgets('Driver app boots to phone authentication when no session exists',
      (tester) async {
    await tester.pumpWidget(const FlashXDriverApp());
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));
    expect(find.text('Sẵn sàng nhận việc cùng FlashX.'), findsOneWidget);
    expect(find.text('Số điện thoại'), findsOneWidget);
    expect(find.text('Vào FlashX Driver'), findsOneWidget);
  });
}
